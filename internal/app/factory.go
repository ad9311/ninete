package app

import (
	"testing"
)

// FactoryConfig loads the application configuration for tests and fails the
// test immediately if loading the config returns an error.
func FactoryConfig(t *testing.T) *Config {
	t.Helper()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("failed to create application config: %v", err)
	}

	return config
}
