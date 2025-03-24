package conf

import (
	"context"
	"fmt"
)

type valueMap struct {
	providers []Provider
	mustLoad  bool
}

var _ ValueMap = (*valueMap)(nil)

// GetInteger retrieves an integer value for the given key.
// Returns 0 if the key is not found or the value cannot be converted to an integer.
func (v *valueMap) GetInt(key string) int {
	value, ok := v.get(key)
	if !ok {
		return 0
	}
	return convertToInt(value)
}

// GetBool retrieves a boolean value for the given key.
// Returns false if the key is not found or the value cannot be converted to a boolean.
func (v *valueMap) GetBool(key string) bool {
	value, ok := v.get(key)
	if !ok {
		return false
	}
	return convertToBool(value)
}

// String retrieves a string value for the given key.
// Returns an empty string if the key is not found.
func (v *valueMap) GetString(key string) string {
	value, ok := v.get(key)
	if ok {
		return value
	}
	return ""
}

// GetBytes retrieves a byte slice value for the given key.
// Returns nil if the key is not found.
func (v *valueMap) GetBytes(key string) []byte {
	value, ok := v.get(key)
	if !ok {
		return nil
	}
	return []byte(value)
}

// MustGetInteger retrieves an integer value for the given key.
// Panics if the key is not found or the value cannot be converted to an integer.
func (v *valueMap) MustGetInt(key string) int {
	value, ok := v.get(key)
	if !ok {
		panic(fmt.Sprintf("key not found: %s", key))
	}
	return convertToInt(value)
}

// MustGetBool retrieves a boolean value for the given key.
// Panics if the key is not found or the value cannot be converted to a boolean.
func (v *valueMap) MustGetBool(key string) bool {
	value, ok := v.get(key)
	if !ok {
		panic(fmt.Sprintf("key not found: %s", key))
	}
	return convertToBool(value)
}

// MustGetString retrieves a string value for the given key.
// Panics if the key is not found.
func (v *valueMap) MustGetString(key string) string {
	value, ok := v.get(key)
	if !ok {
		panic(fmt.Sprintf("key not found: %s", key))
	}
	return value
}

// MustGetBytes retrieves a byte slice value for the given key.
// Panics if the key is not found.
func (v *valueMap) MustGetBytes(key string) []byte {
	value, ok := v.get(key)
	if !ok {
		panic(fmt.Sprintf("key not found: %s", key))
	}
	return []byte(value)
}

// get retrieves a configuration value for the given key from all providers.
// Returns the value and a boolean indicating if the key was found.
func (v *valueMap) get(key string) (string, bool) {
	for _, provider := range v.providers {
		v, ok := provider.Lookup(key)
		if ok {
			return v, ok
		}
	}

	return "", false
}

// New initializes the configuration system with the provided options.
// It processes each provider in the correct order and pulls their values.
// Returns an error if any provider fails to pull its values.
func New(ctx context.Context, opts ...Option) ValueMap {
	// Apply options to create configuration
	valueMap := newConfig(opts...)

	for i, j := 0, len(valueMap.providers)-1; i < j; i, j = i+1, j-1 {
		valueMap.providers[i], valueMap.providers[j] = valueMap.providers[j], valueMap.providers[i]
	}

	envProvider := []Provider{valueMap.providers[len(valueMap.providers)-1]}
	// Process each loader using the providers slice
	for _, provider := range valueMap.providers {
		if err := provider.Load(ctx, envProvider); err != nil {
			if valueMap.mustLoad {
				panic(fmt.Errorf("failed to pull provider: %w", err))
			}
		}
	}

	return valueMap
}
