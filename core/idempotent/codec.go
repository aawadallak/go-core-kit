package idempotent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"

	jsoniter "github.com/json-iterator/go"
	"google.golang.org/protobuf/proto"
)

// Encoder defines a function type that converts any payload into a byte slice.
// Implementations should handle type conversion and serialization.
type Encoder func(payload any) ([]byte, error)

// Decoder defines a function type that converts a payload into a specified type T.
// The type parameter T allows for generic decoding into user-defined structures.
type Decoder[T any] func(payload any) (T, error)

// NewJSONDecoder creates a decoder that unmarshals JSON data into type T.
// It uses jsoniter for faster JSON processing compatible with the standard library.
//
// Type Parameters:
//
//	T - The target type to unmarshal the JSON into
//
// Returns:
//
//	Decoder[T] - A function that decodes JSON byte slices into type T
//
// Errors:
//
//	ErrUnsupportedType - If the payload is not a []byte
//	jsoniter-specific errors - If JSON unmarshaling fails (e.g., malformed JSON)
func NewJSONDecoder[T any]() Decoder[T] {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return func(payload any) (T, error) {
		var target T
		// Verify payload is a byte slice, as JSON unmarshaling requires bytes
		val, ok := payload.([]byte)
		if !ok {
			return target, ErrUnsupportedType
		}
		// Unmarshal JSON bytes into the target type
		if err := json.Unmarshal(val, &target); err != nil {
			return target, err
		}
		return target, nil
	}
}

// NewJSONEncoder creates an encoder that marshals any payload to JSON bytes.
// It uses jsoniter for efficient JSON serialization.
//
// Returns:
//
//	Encoder - A function that encodes any payload to JSON []byte
//
// Errors:
//
//	jsoniter-specific errors - If JSON marshaling fails (e.g., unsupported types)
func NewJSONEncoder() Encoder {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return func(payload any) ([]byte, error) {
		// Marshal the payload directly to JSON
		return json.Marshal(payload)
	}
}

// NewGzipJSONDecoder creates a decoder that decompresses gzip-compressed JSON data
// and unmarshals it into type T. Useful for handling compressed payloads.
//
// Type Parameters:
//
//	T - The target type to unmarshal the decompressed JSON into
//
// Returns:
//
//	Decoder[T] - A function that decodes gzip-compressed JSON into type T
//
// Errors:
//
//	ErrUnsupportedType - If the payload is not a []byte
//	gzip-specific errors - If decompression fails (e.g., invalid gzip data)
//	json-specific errors - If JSON unmarshaling fails (e.g., malformed JSON)
func NewGzipJSONDecoder[T any]() Decoder[T] {
	return func(payload any) (T, error) {
		var target T

		// Ensure payload is a byte slice for gzip decompression
		val, ok := payload.([]byte)
		if !ok {
			return target, ErrUnsupportedType
		}

		// Initialize a reader for the compressed data
		buf := bytes.NewReader(val)
		// Create a gzip reader to decompress the data
		gz, err := gzip.NewReader(buf)
		if err != nil {
			return target, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gz.Close() // Ensure the gzip reader is closed to free resources

		// Read all decompressed data into a byte slice
		decompressed, err := io.ReadAll(gz)
		if err != nil {
			return target, fmt.Errorf("failed to decompress gzip data: %w", err)
		}

		// Unmarshal the decompressed JSON into the target type
		if err := json.Unmarshal(decompressed, &target); err != nil {
			return target, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}

		return target, nil
	}
}

// NewGzipJSONEncoder creates an encoder that marshals a payload to JSON and compresses
// it using gzip. Ideal for reducing payload size in network transmissions.
//
// Returns:
//
//	Encoder - A function that encodes any payload to gzip-compressed JSON []byte
//
// Errors:
//
//	json-specific errors - If JSON marshaling fails
//	gzip-specific errors - If compression fails (rare, typically I/O errors)
func NewGzipJSONEncoder() Encoder {
	return func(payload any) ([]byte, error) {
		// First, marshal the payload to JSON
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON: %w", err)
		}

		// Prepare a buffer to hold the compressed data
		var buf bytes.Buffer
		// Create a gzip writer targeting the buffer
		gz := gzip.NewWriter(&buf)
		// Write the JSON data to the gzip writer for compression
		if _, err := gz.Write(jsonData); err != nil {
			return nil, fmt.Errorf("failed to compress data: %w", err)
		}
		// Close the gzip writer to finalize compression
		if err := gz.Close(); err != nil {
			return nil, fmt.Errorf("failed to close gzip writer: %w", err)
		}

		// Return the compressed bytes
		return buf.Bytes(), nil
	}
}

// NewProtoDecoder creates a decoder that unmarshals Protocol Buffer data into type T.
// The type T must implement proto.Message.
//
// Type Parameters:
//
//	T - The target type, must implement proto.Message
//
// Returns:
//
//	Decoder[T] - A function that decodes Protocol Buffer bytes into type T
//
// Errors:
//
//	ErrUnsupportedType - If the payload is not a []byte
//	proto-specific errors - If unmarshaling fails (e.g., invalid proto data)
func NewProtoDecoder[T proto.Message]() Decoder[T] {
	return func(payload any) (T, error) {
		var target T
		// Verify payload is a byte slice for proto unmarshaling
		val, ok := payload.([]byte)
		if !ok {
			return target, ErrUnsupportedType
		}
		// Unmarshal Protocol Buffer data into the target
		if err := proto.Unmarshal(val, target); err != nil {
			return target, fmt.Errorf("failed to unmarshal proto: %w", err)
		}
		return target, nil
	}
}

// NewProtoEncoder creates an encoder that marshals a payload to Protocol Buffer format.
// The payload must implement proto.Message.
//
// Returns:
//
//	Encoder - A function that encodes proto.Message to []byte
//
// Errors:
//
//	ErrUnsupportedType - If the payload does not implement proto.Message
//	proto-specific errors - If marshaling fails
func NewProtoEncoder() Encoder {
	return func(payload any) ([]byte, error) {
		// Ensure payload implements proto.Message
		msg, ok := payload.(proto.Message)
		if !ok {
			return nil, ErrUnsupportedType
		}
		// Marshal the message to Protocol Buffer format
		data, err := proto.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal proto: %w", err)
		}
		return data, nil
	}
}
