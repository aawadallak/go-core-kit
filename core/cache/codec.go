package cache

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
)

var (
	// ErrInvalidTargetType is returned when the target type is not compatible with the decoder
	ErrInvalidTargetType = errors.New("target must be a byte slice for pure gzip decoding")

	// ErrInvalidPayloadType is returned when the payload type is not compatible with the encoder
	ErrInvalidPayloadType = errors.New("payload must be a byte slice for pure gzip encoding")
)

// Decoder is a function type that decodes a byte payload into a target value.
// It is used to transform cached data from its stored format back into its original type.
type Decoder func(payload []byte, target any) error

// Encoder is a function type that encodes a value into a byte payload.
// It is used to transform data into a format suitable for storage in the cache.
type Encoder func(payload any) ([]byte, error)

// NewDecoderJSON creates a new JSON decoder that unmarshals JSON data into the target value.
// This decoder is suitable for data that was stored as JSON in the cache.
func NewDecoderJSON() Decoder {
	return func(payload []byte, target any) error {
		if err := json.Unmarshal(payload, &target); err != nil {
			return err
		}

		return nil
	}
}

// NewEncoderJSON creates a new JSON encoder that marshals data into JSON format.
// This encoder is suitable for storing data as JSON in the cache.
func NewEncoderJSON() Encoder {
	return func(payload any) ([]byte, error) {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		return body, nil
	}
}

// NewEncoderGzipJSON creates a new encoder that combines JSON marshaling with GZIP compression.
// This encoder is useful for storing large JSON payloads in a compressed format to save space.
// The process involves:
// 1. Converting the payload to JSON
// 2. Compressing the JSON data using GZIP
func NewEncoderGzipJSON() Encoder {
	return func(payload any) ([]byte, error) {
		// Step 1: Marshal to JSON
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		// Step 2: Compress with GZIP
		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		_, err = gzWriter.Write(jsonData)
		if err != nil {
			return nil, err
		}
		if err := gzWriter.Close(); err != nil { // Close to flush the compression
			return nil, err
		}

		return buf.Bytes(), nil
	}
}

// NewDecoderGzip creates a new decoder that decompresses GZIP-compressed data.
// This decoder is designed to work with raw byte slices and expects the target
// to be a pointer to a byte slice (*[]byte).
func NewDecoderGzip() Decoder {
	return func(payload []byte, target any) error {
		reader := bytes.NewReader(payload)
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		defer gzReader.Close()

		data, err := io.ReadAll(gzReader)
		if err != nil {
			return err
		}

		// Assuming target is a []byte
		if b, ok := target.(*[]byte); ok {
			*b = data
			return nil
		}
		return ErrInvalidTargetType
	}
}

// NewEncoderGzip creates a new encoder that compresses data using GZIP.
// This encoder expects the input payload to be a byte slice and compresses it directly.
// It's useful for compressing raw binary data without JSON transformation.
func NewEncoderGzip() Encoder {
	return func(payload any) ([]byte, error) {
		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)

		if b, ok := payload.([]byte); ok {
			_, err := gzWriter.Write(b)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, ErrInvalidPayloadType
		}

		if err := gzWriter.Close(); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
}

// NewDecoderGzipJSON creates a new decoder that combines GZIP decompression with JSON unmarshaling.
// This decoder is designed to work with data that was encoded using NewEncoderGzipJSON.
// The process involves:
// 1. Decompressing the GZIP data
// 2. Unmarshaling the decompressed JSON into the target value
func NewDecoderGzipJSON() Decoder {
	return func(payload []byte, target any) error {
		// Step 1: Decompress GZIP
		reader := bytes.NewReader(payload)
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		defer gzReader.Close()

		// Read decompressed data
		decompressed, err := io.ReadAll(gzReader)
		if err != nil {
			return err
		}

		// Step 2: Unmarshal JSON into target
		if err := json.Unmarshal(decompressed, target); err != nil {
			return err
		}

		return nil
	}
}

// NewSmartEncoder creates an encoder that intelligently chooses between JSON and GZIP+JSON encoding
// based on the size of the payload. This helps optimize storage and performance by:
// - Using simple JSON encoding for small payloads (below threshold)
// - Using GZIP+JSON encoding for large payloads (at or above threshold)
// The encoder avoids double marshaling by reusing the JSON data for compression.
func NewSmartEncoder(threshold int) Encoder {
	return func(payload any) ([]byte, error) {
		// Marshal to JSON once
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		// Check size against threshold
		if len(jsonData) < threshold {
			return jsonData, nil // Return uncompressed JSON
		}

		// Compress with GZIP if above threshold
		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		_, err = gzWriter.Write(jsonData)
		if err != nil {
			return nil, err
		}
		if err := gzWriter.Close(); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}
}
