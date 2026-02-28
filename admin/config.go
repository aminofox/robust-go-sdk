package admin

import (
	"fmt"
	"time"
)

// Config holds configuration for the admin client
type Config struct {
	// Brokers is a list of broker addresses (host:port)
	Brokers []string

	// ClientID is a unique identifier for this client
	ClientID string

	// RequestTimeout is the timeout for requests
	RequestTimeout time.Duration

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// MaxConnections is the maximum number of connections in the pool
	MaxConnections int
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		Brokers:        []string{"localhost:8686"},
		ClientID:       "robust-admin",
		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		MaxConnections: 5,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.Brokers) == 0 {
		return fmt.Errorf("at least one broker required")
	}

	if c.ClientID == "" {
		c.ClientID = "robust-admin"
	}

	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 30 * time.Second
	}

	if c.MaxRetries < 0 {
		c.MaxRetries = 3
	}

	if c.MaxConnections <= 0 {
		c.MaxConnections = 5
	}

	return nil
}
