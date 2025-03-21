package conf

import "context"

// ScanFunc is a function type used for scanning key-value pairs.
// It takes a key and value string as parameters.
type ScanFunc func(key string, value string)

// ValueMap defines the interface for retrieving configuration values.
// It provides methods to get configuration values of different types.
type ValueMap interface {
	// GetInteger retrieves an integer value for the given key.
	// Returns 0 if the key is not found or the value cannot be converted to an integer.
	GetInteger(key string) int

	// GetBoolean retrieves a boolean value for the given key.
	// Returns false if the key is not found or the value cannot be converted to a boolean.
	GetBoolean(key string) bool

	// GetText retrieves a string value for the given key.
	// Returns an empty string if the key is not found.
	GetText(key string) string

	// GetBytes retrieves a byte slice value for the given key.
	// Returns nil if the key is not found.
	GetBytes(key string) []byte

	// MustGetInteger retrieves an integer value for the given key.
	// Panics if the key is not found or the value cannot be converted to an integer.
	MustGetInteger(key string) int

	// MustGetBoolean retrieves a boolean value for the given key.
	// Panics if the key is not found or the value cannot be converted to a boolean.
	MustGetBoolean(key string) bool

	// MustGetText retrieves a string value for the given key.
	// Panics if the key is not found.
	MustGetText(key string) string

	// MustGetBytes retrieves a byte slice value for the given key.
	// Panics if the key is not found.
	MustGetBytes(key string) []byte
}

// Provider defines the interface for configuration providers.
// A provider is responsible for pulling and looking up configuration values.
type Provider interface {
	// Load retrieves configuration values from the provider.
	// The others parameter contains other providers that may be used for dependency resolution.
	// Returns an error if the load operation fails.
	Load(ctx context.Context, others []Provider) error

	// Lookup retrieves a configuration value for the given key.
	// Returns the value and a boolean indicating if the key was found.
	Lookup(key string) (string, bool)

	// Scan iterates over all key-value pairs in the provider.
	// The provided function is called for each key-value pair.
	Scan(fn ScanFunc)
}
