package cache

// options represents the configurable options for a cache instance.
// It holds settings such as encoding/decoding mechanisms and concurrency safety.
type options struct {
	encoder  Encoder // Encoder for serializing cache values
	decoder  Decoder // Decoder for deserializing cache values
	useMutex bool    // Flag to enable mutex-based concurrency protection
}

// Option is a function type that configures a cache instance's options.
// It modifies the options struct to customize the cache behavior.
type Option func(*options)

// WithEncoder returns an Option that sets the encoder for the cache.
// The encoder is responsible for serializing values before storage.
//
// Parameters:
//
//	encoder - The Encoder implementation to use for serialization
//
// Returns:
//
//	Option - A function that sets the encoder in the options struct
func WithEncoder(encoder Encoder) Option {
	return func(o *options) {
		o.encoder = encoder
	}
}

// WithDecoder returns an Option that sets the decoder for the cache.
// The decoder is responsible for deserializing values retrieved from the cache.
//
// Parameters:
//
//	decoder - The Decoder implementation to use for deserialization
//
// Returns:
//
//	Option - A function that sets the decoder in the options struct
func WithDecoder(decoder Decoder) Option {
	return func(o *options) {
		o.decoder = decoder
	}
}

// WithConcurrencySafety returns an Option that enables mutex-based concurrency protection.
// When set, the cache will use a mutex to ensure thread-safe operations.
//
// Returns:
//
//	Option - A function that enables the useMutex flag in the options struct
func WithConcurrencySafety() Option {
	return func(o *options) {
		o.useMutex = true
	}
}

// newOption creates and configures a new options instance with the provided settings.
// It applies each Option function to the default options in sequence.
//
// Parameters:
//
//	opts - Variable number of Option functions to configure the options
//
// Returns:
//
//	*options - Pointer to the configured options struct
func newOption(opts ...Option) *options {
	opt := &options{} // Initialize with default (zero) values

	// Apply each provided option to customize the configuration
	for _, fn := range opts {
		fn(opt)
	}

	return opt
}
