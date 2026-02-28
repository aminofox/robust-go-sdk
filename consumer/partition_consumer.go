package consumer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aminofox/robust-go-sdk/internal/connection"
	"github.com/aminofox/robust-go-sdk/internal/protocol"
	"github.com/aminofox/robust-go-sdk/pkg/message"
	"github.com/google/uuid"
)

// PartitionConsumer consumes messages from a specific topic partition
type PartitionConsumer interface {
	// Messages returns the channel for receiving messages
	Messages() <-chan *message.Message

	// Errors returns the channel for errors
	Errors() <-chan error

	// Close closes the consumer
	Close() error

	// Offset returns the current offset
	Offset() int64
}

// partitionConsumer implements PartitionConsumer
type partitionConsumer struct {
	topic     string
	partition int32
	config    *Config
	pool      *connection.Pool

	messages chan *message.Message
	errors   chan error
	stopCh   chan struct{}

	offset int64
	closed bool
}

// NewPartitionConsumer creates a new partition consumer
func NewPartitionConsumer(topic string, partition int32, config *Config) (PartitionConsumer, error) {
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

	pc := &partitionConsumer{
		topic:     topic,
		partition: partition,
		config:    config,
		pool:      pool,
		messages:  make(chan *message.Message, 100),
		errors:    make(chan error, 10),
		stopCh:    make(chan struct{}),
		offset:    0, // Start from beginning
	}

	// Start polling goroutine
	go pc.poll()

	return pc, nil
}

// poll continuously fetches messages from the broker
func (pc *partitionConsumer) poll() {
	ticker := time.NewTicker(pc.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pc.stopCh:
			return
		case <-ticker.C:
			if err := pc.fetch(); err != nil {
				select {
				case pc.errors <- err:
				case <-pc.stopCh:
					return
				default: // Drop error if channel full
				}
			}
		}
	}
}

// fetch fetches messages from the broker
func (pc *partitionConsumer) fetch() error {
	conn, err := pc.pool.Get()
	if err != nil {
		return err
	}

	// Create fetch request
	fetchReq := &protocol.FetchReq{
		Topic:       pc.topic,
		Partition:   pc.partition,
		Offset:      pc.offset,
		MaxBytes:    pc.config.FetchMaxBytes,
		MaxMessages: pc.config.FetchMaxMessages,
	}

	requestID := uuid.New().String()
	payloadBytes, err := json.Marshal(fetchReq)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req := &protocol.Request{
		Type:      protocol.FetchRequest,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	// Send request
	resp, err := conn.Send(req)
	if err != nil {
		return err
	}

	// Check response
	if !resp.Success {
		return fmt.Errorf("fetch failed: %s", resp.Error)
	}

	// Decode response
	var fetchResp protocol.FetchResp
	if err := json.Unmarshal(resp.Payload, &fetchResp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if fetchResp.Error != "" {
		return fmt.Errorf("fetch error: %s", fetchResp.Error)
	}

	// Send messages to channel
	for _, msg := range fetchResp.Messages {
		select {
		case pc.messages <- msg:
			// Update offset for next fetch
			if msg.Offset >= pc.offset {
				pc.offset = msg.Offset + 1
			}
		case <-pc.stopCh:
			return nil
		}
	}

	return nil
}

// Messages returns the messages channel
func (pc *partitionConsumer) Messages() <-chan *message.Message {
	return pc.messages
}

// Errors returns the errors channel
func (pc *partitionConsumer) Errors() <-chan error {
	return pc.errors
}

// Offset returns the current offset
func (pc *partitionConsumer) Offset() int64 {
	return pc.offset
}

// Close closes the consumer
func (pc *partitionConsumer) Close() error {
	if pc.closed {
		return nil
	}

	pc.closed = true
	close(pc.stopCh)
	close(pc.messages)
	close(pc.errors)
	return pc.pool.Close()
}
