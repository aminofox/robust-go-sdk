package errors

import (
	"errors"
	"fmt"
)

// Common SDK errors
var (
	ErrConnectionFailed   = errors.New("connection failed")
	ErrTopicNotFound      = errors.New("topic not found")
	ErrPartitionNotFound  = errors.New("partition not found")
	ErrOffsetOutOfRange   = errors.New("offset out of range")
	ErrMessageTooLarge    = errors.New("message too large")
	ErrTimeout            = errors.New("request timeout")
	ErrBrokerNotAvailable = errors.New("broker not available")
	ErrInvalidConfig      = errors.New("invalid configuration")
	ErrClosed             = errors.New("client closed")
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
	ErrInvalidResponse    = errors.New("invalid response")
	ErrGroupNotFound      = errors.New("consumer group not found")
)

// ProducerError represents an error that occurred during message production
type ProducerError struct {
	Topic     string
	Partition int32
	Offset    int64
	Err       error
}

func (e *ProducerError) Error() string {
	return fmt.Sprintf("producer error on %s[%d] offset=%d: %v",
		e.Topic, e.Partition, e.Offset, e.Err)
}

func (e *ProducerError) Unwrap() error {
	return e.Err
}

// ConsumerError represents an error that occurred during message consumption
type ConsumerError struct {
	Topic     string
	Partition int32
	Offset    int64
	Err       error
}

func (e *ConsumerError) Error() string {
	return fmt.Sprintf("consumer error on %s[%d] offset=%d: %v",
		e.Topic, e.Partition, e.Offset, e.Err)
}

func (e *ConsumerError) Unwrap() error {
	return e.Err
}

// ConnectionError represents a connection-related error
type ConnectionError struct {
	Broker string
	Err    error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error to %s: %v", e.Broker, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}
