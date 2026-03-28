package idem

import "github.com/vmihailenco/msgpack/v5"

type MsgPackCodec struct{}

func (MsgPackCodec) Marshal(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (MsgPackCodec) Unmarshal(data []byte, target any) error {
	return msgpack.Unmarshal(data, target)
}
