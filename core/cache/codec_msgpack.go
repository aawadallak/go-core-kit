package cache

import (
	"github.com/vmihailenco/msgpack/v5"
)

// NewEncoderMsgPack returns a lightweight binary encoder using MessagePack
func NewEncoderMsgPack() Encoder {
	return msgpack.Marshal
}

// NewDecoderMsgPack returns a lightweight binary decoder using MessagePack
func NewDecoderMsgPack() Decoder {
	return msgpack.Unmarshal
}
