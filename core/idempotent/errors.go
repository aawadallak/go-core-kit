package idempotent

import "errors"

// ErrUnsupportedType is returned when a type is not supported by the encoder or decoder.
var ErrUnsupportedType = errors.New("unsupported type")
