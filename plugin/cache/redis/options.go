// Package redis provides Redis caching functionality with configurable options.
package redis

import (
	"errors"
	"time"

	"crypto/tls"

	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidAddress   = errors.New("invalid Redis address")
	ErrInvalidPoolSize  = errors.New("invalid pool size")
	ErrInvalidIdleConns = errors.New("min idle connections cannot be greater than pool size")
)

// options specifies the configuration options for a Redis provider.
type options struct {
	redisOptions *redis.Options
}

// Option is a function that configures a Redis provider option.
type Option func(*options)

// WithAddress sets the Redis server address option.
func WithAddress(addr string) Option {
	return func(o *options) {
		o.redisOptions.Addr = addr
	}
}

// WithPassword sets the Redis server password option.
func WithPassword(password string) Option {
	return func(o *options) {
		o.redisOptions.Password = password
	}
}

// WithDB sets the Redis database number option.
func WithDB(db int) Option {
	return func(o *options) {
		o.redisOptions.DB = db
	}
}

// WithMaxRetry sets the maximum number of retry attempts option.
func WithMaxRetry(attempts uint) Option {
	return func(o *options) {
		o.redisOptions.MaxRetries = int(attempts) //nolint:gosec // retry attempts are inherently small values
	}
}

// WithRetryBackoff sets the retry backoff time range options.
func WithRetryBackoff(minBackoff, maxBackoff time.Duration) Option {
	return func(o *options) {
		o.redisOptions.MinRetryBackoff = minBackoff
		o.redisOptions.MaxRetryBackoff = maxBackoff
	}
}

// WithPoolSize sets the connection pool size option.
func WithPoolSize(size uint) Option {
	return func(o *options) {
		o.redisOptions.PoolSize = int(size) //nolint:gosec // pool size is inherently a small value
	}
}

// WithMinIdleConns sets the minimum number of idle connections option.
func WithMinIdleConns(size uint) Option {
	return func(o *options) {
		o.redisOptions.MinIdleConns = int(size) //nolint:gosec // idle conns is inherently a small value
	}
}

// WithMaxIdleConns sets the maximum number of idle connections option.
func WithMaxIdleConns(size uint) Option {
	return func(o *options) {
		o.redisOptions.MaxIdleConns = int(size) //nolint:gosec // small values
	}
}

// WithTimeouts sets all timeout options.
func WithTimeouts(dial, read, write, pool time.Duration) Option {
	return func(o *options) {
		o.redisOptions.DialTimeout = dial
		o.redisOptions.ReadTimeout = read
		o.redisOptions.WriteTimeout = write
		o.redisOptions.PoolTimeout = pool
	}
}

// WithTLS enables TLS connection with the given configuration.
func WithTLS(config *tls.Config) Option {
	return func(o *options) {
		o.redisOptions.TLSConfig = config
	}
}

// newOptions creates a new options instance with the given options.
// It uses default options as a base and applies any provided options.
func newOptions(opts ...Option) *options {
	o := &options{
		redisOptions: newDefaultOptions(),
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// validate checks if the options are valid.
func (o *options) validate() error {
	if o.redisOptions.Addr == "" {
		return ErrInvalidAddress
	}
	if o.redisOptions.PoolSize < 1 {
		return ErrInvalidPoolSize
	}
	if o.redisOptions.MinIdleConns > o.redisOptions.PoolSize {
		return ErrInvalidIdleConns
	}
	return nil
}
