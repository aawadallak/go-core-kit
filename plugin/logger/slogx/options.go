package slogx

type options struct {
	addSource bool
}

type Option func(*options)

func WithAddSource() Option {
	return func(o *options) {
		o.addSource = true
	}
}

func newOptions(opts ...Option) *options {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}
