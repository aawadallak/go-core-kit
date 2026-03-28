package msgpack_test

import (
	"testing"

	"github.com/aawadallak/go-core-kit/plugin/cache/msgpack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoderDecoderRoundTrip_String(t *testing.T) {
	encode := msgpack.NewEncoder()
	decode := msgpack.NewDecoder()

	original := "hello, msgpack"

	data, err := encode(original)
	require.NoError(t, err)

	var decoded string
	err = decode(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original, decoded)
}

func TestEncoderDecoderRoundTrip_Int(t *testing.T) {
	encode := msgpack.NewEncoder()
	decode := msgpack.NewDecoder()

	original := 42

	data, err := encode(original)
	require.NoError(t, err)

	var decoded int
	err = decode(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original, decoded)
}

func TestEncoderDecoderRoundTrip_Struct(t *testing.T) {
	type sample struct {
		Name  string `msgpack:"name"`
		Value int    `msgpack:"value"`
	}

	encode := msgpack.NewEncoder()
	decode := msgpack.NewDecoder()

	original := sample{Name: "test", Value: 99}

	data, err := encode(original)
	require.NoError(t, err)

	var decoded sample
	err = decode(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original, decoded)
}
