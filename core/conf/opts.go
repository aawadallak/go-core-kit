package conf

// Option is a function type that modifies a config instance.
type Option func(*config)

// config holds the configuration state including all providers.
type config struct {
	providers []Provider
}

// WithProvider adds a new provider to the configuration.
// Returns an Option function that can be used to modify a config instance.
func WithProvider(provider Provider) Option {
	return func(c *config) {
		c.providers = append(c.providers, provider)
	}
}

// newConfig creates a new config instance with the provided options.
// It initializes with a default provider and applies any additional options.
// Returns a pointer to the new config instance.
func newConfig(opts ...Option) *config {
	c := &config{
		providers: []Provider{NewProvider()},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
