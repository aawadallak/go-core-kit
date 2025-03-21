package conf

import (
	"context"
	"fmt"
)

// GetInteger retrieves an integer value for the given key.
// Returns 0 if the key is not found or the value cannot be converted to an integer.
func (c *config) GetInteger(key string) int {
	value, ok := c.get(key)
	if !ok {
		return 0
	}
	return convertToInt(value)
}

// GetBoolean retrieves a boolean value for the given key.
// Returns false if the key is not found or the value cannot be converted to a boolean.
func (c *config) GetBoolean(key string) bool {
	value, ok := c.get(key)
	if !ok {
		return false
	}
	return convertToBool(value)
}

// GetText retrieves a string value for the given key.
// Returns an empty string if the key is not found.
func (c *config) GetText(key string) string {
	value, ok := c.get(key)
	if ok {
		return value
	}
	return ""
}

// GetBytes retrieves a byte slice value for the given key.
// Returns nil if the key is not found.
func (c *config) GetBytes(key string) []byte {
	value, ok := c.get(key)
	if !ok {
		return nil
	}
	return []byte(value)
}

// MustGetInteger retrieves an integer value for the given key.
// Panics if the key is not found or the value cannot be converted to an integer.
func (c *config) MustGetInteger(key string) int {
	value, ok := c.get(key)
	if !ok {
		panic(fmt.Sprintf("key not found: %s", key))
	}
	return convertToInt(value)
}

// MustGetBoolean retrieves a boolean value for the given key.
// Panics if the key is not found or the value cannot be converted to a boolean.
func (c *config) MustGetBoolean(key string) bool {
	value, ok := c.get(key)
	if !ok {
		panic(fmt.Sprintf("key not found: %s", key))
	}
	return convertToBool(value)
}

// MustGetText retrieves a string value for the given key.
// Panics if the key is not found.
func (c *config) MustGetText(key string) string {
	value, ok := c.get(key)
	if !ok {
		panic(fmt.Sprintf("key not found: %s", key))
	}
	return value
}

// MustGetBytes retrieves a byte slice value for the given key.
// Panics if the key is not found.
func (c *config) MustGetBytes(key string) []byte {
	value, ok := c.get(key)
	if !ok {
		panic(fmt.Sprintf("key not found: %s", key))
	}
	return []byte(value)
}

// get retrieves a configuration value for the given key from all providers.
// Returns the value and a boolean indicating if the key was found.
func (c *config) get(key string) (string, bool) {
	for _, provider := range c.providers {
		v, ok := provider.Lookup(key)
		if ok {
			return v, ok
		}
	}
	return "", false
}

// Init initializes the configuration system with the provided options.
// It processes each provider in the correct order and pulls their values.
// Returns an error if any provider fails to pull its values.
func Init(ctx context.Context, opts ...Option) error {
	// Apply options to create configuration
	cfg := newConfig(opts...)

	// Convert loaders to providers once and reverse
	providers := make([]Provider, len(cfg.providers))
	if len(providers) > 1 {
		providers = reverseProviders(providers)
	}

	// Process each loader using the providers slice
	for _, provider := range cfg.providers {
		if err := provider.Load(ctx, providers); err != nil {
			return fmt.Errorf("failed to pull provider: %w", err)
		}
	}

	return nil
}

// reverseProviders reverses a slice of providers efficiently.
// Returns a new slice with the providers in reverse order.
func reverseProviders(providers []Provider) []Provider {
	reversed := make([]Provider, len(providers))
	for i, j := 0, len(providers)-1; j >= 0; i, j = i+1, j-1 {
		reversed[i] = providers[j]
	}
	return reversed
}
