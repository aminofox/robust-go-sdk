package producer

import (
	"time"
)

// Config configures a producer
type Config struct {
	// Brokers is the list of broker addresses
	Brokers []string

	// ClientID identifies this producer
	ClientID string

	// RequestTimeout is the timeout for produce requests
	RequestTimeout time.Duration

	// RetryBackoff is the time to wait before retrying
	RetryBackoff time.Duration

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// MaxConnections is the maximum number of connections in the pool
	MaxConnections int

	// BatchSize is the number of messages to batch (for async producer)
	BatchSize int

	// BatchTimeout is the timeout for batching (for async producer)
	BatchTimeout time.Duration
}

// DefaultConfig returns a default producer configuration
func DefaultConfig() *Config {
	return &Config{
		Brokers:        []string{"localhost:8686"},
		ClientID:       "robust-producer",
		RequestTimeout: 30 * time.Second,
		RetryBackoff:   100 * time.Millisecond,
		MaxRetries:     3,
		MaxConnections: 5,
		BatchSize:      100,
		BatchTimeout:   10 * time.Millisecond,
	}
}

// Validate validates the producer configuration
func (c *Config) Validate() error {
	if len(c.Brokers) == 0 {
		return ErrInvalidConfig("no brokers specified")
	}

	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 30 * time.Second
	}

	if c.MaxRetries < 0 {
		c.MaxRetries = 0
	}

	if c.MaxConnections <= 0 {
		c.MaxConnections = 5
	}

	return nil
}

// ErrInvalidConfig creates an invalid config error
func ErrInvalidConfig(msg string) error {
	return &ConfigError{Message: msg}
}

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return "invalid producer config: " + e.Message
}
