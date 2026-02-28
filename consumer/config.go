package consumer

import (
	"time"
)

// Config configures a consumer
type Config struct {
	// Brokers is the list of broker addresses
	Brokers []string

	// GroupID for consumer group (optional)
	GroupID string

	// ClientID identifies this consumer
	ClientID string

	// Topics to subscribe
	Topics []string

	// FetchMaxBytes is the maximum bytes to fetch per request
	FetchMaxBytes int

	// FetchMaxMessages is the maximum messages to fetch per request
	FetchMaxMessages int

	// PollInterval is how often to poll for new messages
	PollInterval time.Duration

	// RequestTimeout is the timeout for requests
	RequestTimeout time.Duration

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// MaxConnections is the maximum number of connections in the pool
	MaxConnections int
}

// DefaultConfig returns a default consumer configuration
func DefaultConfig() *Config {
	return &Config{
		Brokers:          []string{"localhost:8686"},
		ClientID:         "robust-consumer",
		FetchMaxBytes:    1048576, // 1MB
		FetchMaxMessages: 100,
		PollInterval:     100 * time.Millisecond,
		RequestTimeout:   30 * time.Second,
		MaxRetries:       3,
		MaxConnections:   5,
	}
}

// Validate validates the consumer configuration
func (c *Config) Validate() error {
	if len(c.Brokers) == 0 {
		return ErrInvalidConfig("no brokers specified")
	}

	if c.FetchMaxBytes <= 0 {
		c.FetchMaxBytes = 1048576
	}

	if c.FetchMaxMessages <= 0 {
		c.FetchMaxMessages = 100
	}

	if c.PollInterval <= 0 {
		c.PollInterval = 100 * time.Millisecond
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
	return "invalid consumer config: " + e.Message
}
