package producer

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aminofox/robust-go-sdk/internal/connection"
	"github.com/aminofox/robust-go-sdk/internal/protocol"
	"github.com/aminofox/robust-go-sdk/pkg/message"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ProducerMessage represents a message with its result
type ProducerMessage struct {
	Message   *message.Message
	Partition int32
	Offset    int64
	Metadata  interface{}
}

// ProducerError represents a producer error
type ProducerError struct {
	Msg      *message.Message
	Err      error
	Metadata interface{}
}

// AsyncProducer sends messages asynchronously
type AsyncProducer interface {
	// Input returns the input channel for messages
	Input() chan<- *message.Message

	// Successes returns the success channel
	Successes() <-chan *ProducerMessage

	// Errors returns the error channel
	Errors() <-chan *ProducerError

	// Close closes the producer
	Close() error
}

// asyncProducer implements AsyncProducer
type asyncProducer struct {
	config    *Config
	pool      *connection.Pool
	input     chan *message.Message
	successes chan *ProducerMessage
	errors    chan *ProducerError
	closed    bool
	logger    *zap.Logger
	wg        sync.WaitGroup
	mu        sync.RWMutex
}

// NewAsyncProducer creates a new async producer
func NewAsyncProducer(config *Config) (AsyncProducer, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	pool, err := connection.NewPool(&connection.PoolConfig{
		Brokers:        config.Brokers,
		Size:           config.MaxConnections,
		ConnectTimeout: config.RequestTimeout,
	})
	if err != nil {
		return nil, err
	}

	logger, _ := zap.NewProduction()

	ap := &asyncProducer{
		config:    config,
		pool:      pool,
		input:     make(chan *message.Message, 1000),
		successes: make(chan *ProducerMessage, 100),
		errors:    make(chan *ProducerError, 100),
		logger:    logger,
	}

	// Start worker goroutines
	workerCount := config.MaxConnections
	for i := 0; i < workerCount; i++ {
		ap.wg.Add(1)
		go ap.worker()
	}

	// Start batcher if batching is enabled
	if config.BatchSize > 1 {
		ap.wg.Add(1)
		go ap.batcher()
	}

	return ap, nil
}

// Input returns the input channel
func (ap *asyncProducer) Input() chan<- *message.Message {
	return ap.input
}

// Successes returns the success channel
func (ap *asyncProducer) Successes() <-chan *ProducerMessage {
	return ap.successes
}

// Errors returns the error channel
func (ap *asyncProducer) Errors() <-chan *ProducerError {
	return ap.errors
}

// worker processes messages from the input channel
func (ap *asyncProducer) worker() {
	defer ap.wg.Done()

	for msg := range ap.input {
		ap.mu.RLock()
		if ap.closed {
			ap.mu.RUnlock()
			return
		}
		ap.mu.RUnlock()

		partition, offset, err := ap.sendMessage(msg)
		if err != nil {
			select {
			case ap.errors <- &ProducerError{
				Msg: msg,
				Err: err,
			}:
			default:
				ap.logger.Warn("error channel full, dropping error",
					zap.Error(err))
			}
			continue
		}

		select {
		case ap.successes <- &ProducerMessage{
			Message:   msg,
			Partition: partition,
			Offset:    offset,
		}:
		default:
			ap.logger.Warn("success channel full, dropping success")
		}
	}
}

// batcher batches messages before sending
func (ap *asyncProducer) batcher() {
	defer ap.wg.Done()

	ticker := time.NewTicker(ap.config.BatchTimeout)
	defer ticker.Stop()

	batch := make([]*message.Message, 0, ap.config.BatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		// Send batch
		for _, msg := range batch {
			partition, offset, err := ap.sendMessage(msg)
			if err != nil {
				select {
				case ap.errors <- &ProducerError{
					Msg: msg,
					Err: err,
				}:
				default:
				}
				continue
			}

			select {
			case ap.successes <- &ProducerMessage{
				Message:   msg,
				Partition: partition,
				Offset:    offset,
			}:
			default:
			}
		}

		batch = batch[:0]
	}

	for {
		select {
		case msg, ok := <-ap.input:
			if !ok {
				flush()
				return
			}

			batch = append(batch, msg)
			if len(batch) >= ap.config.BatchSize {
				flush()
			}

		case <-ticker.C:
			flush()
		}
	}
}

// sendMessage sends a single message
func (ap *asyncProducer) sendMessage(msg *message.Message) (int32, int64, error) {
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}

	conn, err := ap.pool.Get()
	if err != nil {
		return 0, 0, fmt.Errorf("get connection: %w", err)
	}

	produceReq := &protocol.ProduceReq{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Key:       msg.Key,
		Value:     msg.Value,
		Headers:   msg.Headers,
		Timestamp: msg.Timestamp,
	}

	requestID := uuid.New().String()
	payloadBytes, err := json.Marshal(produceReq)
	if err != nil {
		return 0, 0, fmt.Errorf("marshal request: %w", err)
	}

	req := &protocol.Request{
		Type:      protocol.ProduceRequest,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	var lastErr error
	for attempt := 0; attempt <= ap.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(ap.config.RetryBackoff)
		}

		resp, err := conn.Send(req)
		if err != nil {
			lastErr = err
			conn, _ = ap.pool.Get()
			continue
		}

		if !resp.Success {
			lastErr = fmt.Errorf("produce failed: %s", resp.Error)
			continue
		}

		var produceResp protocol.ProduceResp
		if err := json.Unmarshal(resp.Payload, &produceResp); err != nil {
			lastErr = fmt.Errorf("unmarshal response: %w", err)
			continue
		}

		if produceResp.Error != "" {
			lastErr = fmt.Errorf("produce error: %s", produceResp.Error)
			continue
		}

		return produceResp.Partition, produceResp.Offset, nil
	}

	return 0, 0, lastErr
}

// Close closes the producer
func (ap *asyncProducer) Close() error {
	ap.mu.Lock()
	if ap.closed {
		ap.mu.Unlock()
		return nil
	}
	ap.closed = true
	ap.mu.Unlock()

	close(ap.input)
	ap.wg.Wait()

	close(ap.successes)
	close(ap.errors)

	return ap.pool.Close()
}
