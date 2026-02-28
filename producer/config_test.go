package producer

import (
	"testing"
)

func TestConfigDefault(t *testing.T) {
	config := DefaultConfig()

	if len(config.Brokers) == 0 {
		t.Error("DefaultConfig should have brokers")
	}

	if config.RequestTimeout <= 0 {
		t.Error("RequestTimeout should be positive")
	}

	if config.MaxRetries < 0 {
		t.Error("MaxRetries should not be negative")
	}
}

func TestConfigValidation(t *testing.T) {
	config := &Config{
		Brokers: []string{"localhost:8686"},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	emptyConfig := &Config{}
	err = emptyConfig.Validate()
	if err == nil {
		t.Error("Expected error for empty brokers")
	}
}
