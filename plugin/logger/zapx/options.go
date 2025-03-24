package zapx

type options struct {
	enableCaller     bool
	enableStackTrace bool
}

type Option func(*options)

// WithAddCaller enables caller information in the log output
func WithEnableCaller() Option {
	return func(o *options) {
		o.enableCaller = true
	}
}

func WithEnableStackTrace() Option {
	return func(o *options) {
		o.enableStackTrace = true
	}
}

func newOptions(opts ...Option) *options {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}
