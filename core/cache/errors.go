package cache

import "errors"

var (
	// ErrKeyNotFound is returned in Cache.Get and Cache.Delete when the
	// provided key could not be found in cache.
	ErrKeyNotFound error = errors.New("cache: key not found in cache")

	// ErrInvalidCachedValue is returned when the cached value cannot be decoded
	// into the expected type.
	ErrInvalidCachedValue error = errors.New("cache: invalid cached value")

	// ErrInvalidTargetType is returned when the target type is not compatible with the decoder
	ErrInvalidTargetType error = errors.New("cache: target must be a byte slice for pure gzip decoding")

	// ErrInvalidPayloadType is returned when the payload type is not compatible with the encoder
	ErrInvalidPayloadType error = errors.New("cache: payload must be a byte slice for pure gzip encoding")

	// ErrClosed is returned when attempting to use a closed cache
	ErrClosed error = errors.New("cache: closed")
)
