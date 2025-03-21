package cache

// options represents the configurable options for a cache instance.
type options struct {
	encoder  Encoder
	decoder  Decoder
	useMutex bool
}

// Option is a function that sets a configuration option for a cache instance.
type Option func(*options)

func WithEncoder(encoder Encoder) Option {
	return func(o *options) {
		o.encoder = encoder
	}
}

func WithDecoder(decoder Decoder) Option {
	return func(o *options) {
		o.decoder = decoder
	}
}

// WithConcurrencySafety enables mutex-based concurrency protection
func WithConcurrencySafety() Option {
	return func(o *options) {
		o.useMutex = true
	}
}

func newOption(opts ...Option) *options {
	opt := &options{}

	for _, fn := range opts {
		fn(opt)
	}

	return opt
}
