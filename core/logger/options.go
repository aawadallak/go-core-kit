package logger

// Option represents a function that modifies a Config instance.
// It follows the functional options pattern for flexible configuration.
type Option func(*Config)

// Config holds logger configuration settings.
// It stores attributes as key-value pairs and an optional custom Provider.
type Config struct {
	Attributes map[string]any // Key-value pairs to be included in log messages
	Provider   Provider       // Custom logging provider, overrides default if set
}

// WithValue returns an Option that adds or updates a key-value pair in the Config.
// The value can be of any type and will be included in structured log output.
func WithValue(key string, value any) Option {
	return func(c *Config) {
		c.Attributes[key] = value
	}
}

// WithProvider returns an Option that sets a custom Provider in the Config.
// This allows overriding the default logging implementation.
func WithProvider(provider Provider) Option {
	return func(c *Config) {
		c.Provider = provider
	}
}

// NewConfig creates a new Config instance with the specified options applied.
// It initializes an empty Attributes map and applies each option in sequence.
// Returns a pointer to the configured Config instance.
func NewConfig(opts ...Option) *Config {
	cfg := &Config{
		Attributes: make(map[string]any), // Initialize to avoid nil map issues
	}

	// Apply each option to the configuration
	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}
