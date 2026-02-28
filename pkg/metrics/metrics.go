package metrics

import (
	"sync/atomic"
	"time"
)

// ProducerMetrics tracks producer-related metrics
type ProducerMetrics struct {
	MessagesSent   atomic.Uint64
	MessagesFailed atomic.Uint64
	BytesSent      atomic.Uint64
	startTime      time.Time
}

// NewProducerMetrics creates a new ProducerMetrics instance
func NewProducerMetrics() *ProducerMetrics {
	return &ProducerMetrics{
		startTime: time.Now(),
	}
}

// RecordSent records a successfully sent message
func (m *ProducerMetrics) RecordSent(bytes uint64) {
	m.MessagesSent.Add(1)
	m.BytesSent.Add(bytes)
}

// RecordFailed records a failed message send
func (m *ProducerMetrics) RecordFailed() {
	m.MessagesFailed.Add(1)
}

// GetStats returns the current producer statistics
func (m *ProducerMetrics) GetStats() ProducerStats {
	return ProducerStats{
		MessagesSent:   m.MessagesSent.Load(),
		MessagesFailed: m.MessagesFailed.Load(),
		BytesSent:      m.BytesSent.Load(),
		Uptime:         time.Since(m.startTime),
	}
}

// ProducerStats represents producer statistics
type ProducerStats struct {
	MessagesSent   uint64
	MessagesFailed uint64
	BytesSent      uint64
	Uptime         time.Duration
}

// ConsumerMetrics tracks consumer-related metrics
type ConsumerMetrics struct {
	MessagesReceived atomic.Uint64
	BytesReceived    atomic.Uint64
	startTime        time.Time
}

// NewConsumerMetrics creates a new ConsumerMetrics instance
func NewConsumerMetrics() *ConsumerMetrics {
	return &ConsumerMetrics{
		startTime: time.Now(),
	}
}

// RecordReceived records a successfully received message
func (m *ConsumerMetrics) RecordReceived(bytes uint64) {
	m.MessagesReceived.Add(1)
	m.BytesReceived.Add(bytes)
}

// GetStats returns the current consumer statistics
func (m *ConsumerMetrics) GetStats() ConsumerStats {
	return ConsumerStats{
		MessagesReceived: m.MessagesReceived.Load(),
		BytesReceived:    m.BytesReceived.Load(),
		Uptime:           time.Since(m.startTime),
	}
}

// ConsumerStats represents consumer statistics
type ConsumerStats struct {
	MessagesReceived uint64
	BytesReceived    uint64
	Uptime           time.Duration
}

// ConnectionMetrics tracks connection-related metrics
type ConnectionMetrics struct {
	ActiveConnections atomic.Int64
	TotalConnections  atomic.Uint64
	FailedConnections atomic.Uint64
	startTime         time.Time
}

// NewConnectionMetrics creates a new ConnectionMetrics instance
func NewConnectionMetrics() *ConnectionMetrics {
	return &ConnectionMetrics{
		startTime: time.Now(),
	}
}

// RecordConnectionOpened records a new connection
func (m *ConnectionMetrics) RecordConnectionOpened() {
	m.ActiveConnections.Add(1)
	m.TotalConnections.Add(1)
}

// RecordConnectionClosed records a closed connection
func (m *ConnectionMetrics) RecordConnectionClosed() {
	m.ActiveConnections.Add(-1)
}

// RecordConnectionFailed records a failed connection attempt
func (m *ConnectionMetrics) RecordConnectionFailed() {
	m.FailedConnections.Add(1)
}

// GetStats returns the current connection statistics
func (m *ConnectionMetrics) GetStats() ConnectionStats {
	return ConnectionStats{
		ActiveConnections: m.ActiveConnections.Load(),
		TotalConnections:  m.TotalConnections.Load(),
		FailedConnections: m.FailedConnections.Load(),
		Uptime:            time.Since(m.startTime),
	}
}

// ConnectionStats represents connection statistics
type ConnectionStats struct {
	ActiveConnections int64
	TotalConnections  uint64
	FailedConnections uint64
	Uptime            time.Duration
}
