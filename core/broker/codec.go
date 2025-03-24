package broker

// Encoder is a function type that converts a Message into a byte slice.
// It takes a Message as input and returns the encoded bytes along with any error encountered during encoding.
type Encoder func(Message) ([]byte, error)

// Decoder is a function type that converts a byte slice into a generic value.
// It takes a byte slice as input and returns the decoded value as an interface{} along with any error encountered during decoding.
type Decoder func([]byte) (any, error)
