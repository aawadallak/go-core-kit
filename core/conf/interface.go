package conf

import "context"

// ScanFunc defines a function type for scanning configuration key-value pairs.
// It accepts key and value as string parameters.
type ScanFunc func(key string, value string)

// ValueMap specifies an interface for accessing configuration values.
// It offers methods to retrieve values in various data types.
type ValueMap interface {
	// GetInt returns the integer value associated with the specified key.
	// Returns 0 if the key doesn't exist or if conversion to integer fails.
	GetInt(key string) int

	// GetBool returns the boolean value associated with the specified key.
	// Returns false if the key doesn't exist or if conversion to boolean fails.
	GetBool(key string) bool

	// GetString returns the string value associated with the specified key.
	// Returns an empty string if the key doesn't exist.
	GetString(key string) string

	// GetBytes returns the byte slice value associated with the specified key.
	// Returns nil if the key doesn't exist.
	GetBytes(key string) []byte

	// MustGetInt returns the integer value for the specified key.
	// Panics if the key doesn't exist or if conversion to integer fails.
	MustGetInt(key string) int

	// MustGetBool returns the boolean value for the specified key.
	// Panics if the key doesn't exist or if conversion to boolean fails.
	MustGetBool(key string) bool

	// MustGetString returns the string value for the specified key.
	// Panics if the key doesn't exist.
	MustGetString(key string) string

	// MustGetBytes returns the byte slice value for the specified key.
	// Panics if the key doesn't exist.
	MustGetBytes(key string) []byte
}

// Provider defines an interface for configuration providers.
// It handles the retrieval and lookup of configuration values.
type Provider interface {
	// Load fetches configuration values from the provider.
	// Uses context for cancellation/timeout and other providers for dependency resolution.
	// Returns an error if the loading process fails.
	Load(ctx context.Context, others []Provider) error

	// Lookup retrieves the value for a given key.
	// Returns the value as a string and a boolean indicating if the key was found.
	Lookup(key string) (string, bool)

	// Scan iterates through all key-value pairs in the provider.
	// Executes the provided function for each pair encountered.
	Scan(fn ScanFunc)
}
