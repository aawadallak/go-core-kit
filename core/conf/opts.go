package conf

// Option represents a function that modifies a configuration instance.
// It takes a pointer to a valueMap and applies specific configuration changes.
type Option func(*valueMap)

// WithProvider creates an Option that adds a new provider to the configuration.
// The provider will be appended to the existing list of providers used for
// retrieving configuration values.
//
// Parameters:
//
//	provider - The Provider implementation to add to the configuration
//
// Returns:
//
//	An Option function that modifies a config instance by adding the provider
func WithProvider(provider Provider) Option {
	return func(c *valueMap) {
		c.providers = append(c.providers, provider)
	}
}

// WithMustLoad creates an Option that sets the configuration to require
// successful loading of values. When enabled, the configuration will panic
// if values cannot be loaded from any provider.
//
// Returns:
//
//	An Option function that enables the mustLoad flag on a config instance
func WithMustLoad() Option {
	return func(c *valueMap) {
		c.mustLoad = true
	}
}

// newConfig initializes and returns a new configuration instance with the
// specified options. The configuration starts with a default environment
// provider and applies all provided options in the order they are given.
//
// Parameters:
//
//	opts - Variable number of Option functions to apply to the configuration
//
// Returns:
//
//	A pointer to the newly created and configured valueMap instance
func newConfig(opts ...Option) *valueMap {
	c := &valueMap{
		providers: []Provider{newEnvProvider()},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
