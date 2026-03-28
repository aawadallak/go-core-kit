package msgpack

import (
	"github.com/aawadallak/go-core-kit/core/cache"
	"github.com/vmihailenco/msgpack/v5"
)

// NewEncoder returns a lightweight binary encoder using MessagePack.
func NewEncoder() cache.Encoder {
	return msgpack.Marshal
}

// NewDecoder returns a lightweight binary decoder using MessagePack.
func NewDecoder() cache.Decoder {
	return msgpack.Unmarshal
}
