package producer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aminofox/robust-go-sdk/internal/connection"
	"github.com/aminofox/robust-go-sdk/internal/protocol"
	sdkerrors "github.com/aminofox/robust-go-sdk/pkg/errors"
	"github.com/aminofox/robust-go-sdk/pkg/message"
	"github.com/google/uuid"
)

// SyncProducer sends messages synchronously and waits for acknowledgment
type SyncProducer interface {
	// SendMessage sends a single message and returns partition and offset
	SendMessage(msg *message.Message) (partition int32, offset int64, err error)

	// SendMessages sends multiple messages in sequence
	SendMessages(msgs []*message.Message) error

	// Close closes the producer and releases resources
	Close() error
}

// syncProducer implements SyncProducer
type syncProducer struct {
	config *Config
	pool   *connection.Pool
	closed bool
}

// NewSyncProducer creates a new synchronous producer
func NewSyncProducer(config *Config) (SyncProducer, error) {
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

	sp := &syncProducer{
		config: config,
		pool:   pool,
	}

	return sp, nil
}

// SendMessage sends a single message
func (p *syncProducer) SendMessage(msg *message.Message) (int32, int64, error) {
	if p.closed {
		return 0, 0, sdkerrors.ErrClosed
	}

	// Set timestamp if not set
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}

	// Initialize headers if nil
	if msg.Headers == nil {
		msg.Headers = make(map[string]string)
	}

	// Build produce request
	produceReq := &protocol.ProduceReq{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Key:       msg.Key,
		Value:     msg.Value,
		Headers:   msg.Headers,
	}

	// Try to send with retries
	var lastErr error
	for i := 0; i <= p.config.MaxRetries; i++ {
		if i > 0 {
			time.Sleep(p.config.RetryBackoff)
		}

		partition, offset, err := p.sendWithConn(produceReq)
		if err == nil {
			return partition, offset, nil
		}

		lastErr = err
	}

	return 0, 0, &sdkerrors.ProducerError{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Err:       lastErr,
	}
}

// sendWithConn sends a produce request using a connection from the pool
func (p *syncProducer) sendWithConn(produceReq *protocol.ProduceReq) (int32, int64, error) {
	conn, err := p.pool.Get()
	if err != nil {
		return 0, 0, err
	}

	// Create request envelope
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

	// Send request
	resp, err := conn.Send(req)
	if err != nil {
		return 0, 0, err
	}

	// Check response
	if !resp.Success {
		return 0, 0, fmt.Errorf("produce failed: %s", resp.Error)
	}

	// Decode response
	var produceResp protocol.ProduceResp
	if err := json.Unmarshal(resp.Payload, &produceResp); err != nil {
		return 0, 0, fmt.Errorf("unmarshal response: %w", err)
	}

	if produceResp.Error != "" {
		return 0, 0, fmt.Errorf("produce error: %s", produceResp.Error)
	}

	return produceResp.Partition, produceResp.Offset, nil
}

// SendMessages sends multiple messages sequentially
func (p *syncProducer) SendMessages(msgs []*message.Message) error {
	if p.closed {
		return sdkerrors.ErrClosed
	}

	for _, msg := range msgs {
		_, _, err := p.SendMessage(msg)
		if err != nil {
			return err
		}
	}

	return nil
}

// Close closes the producer
func (p *syncProducer) Close() error {
	if p.closed {
		return nil
	}

	p.closed = true
	return p.pool.Close()
}
