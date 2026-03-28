package onepassword

import (
	"context"
	"testing"

	"github.com/aawadallak/go-core-kit/core/conf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type envProvider struct {
	data map[string]string
}

func (e *envProvider) Load(_ context.Context, _ []conf.Provider) error { return nil }
func (e *envProvider) Lookup(key string) (string, bool) {
	v, ok := e.data[key]
	return v, ok
}
func (e *envProvider) Scan(fn conf.ScanFunc) {
	for k, v := range e.data {
		fn(k, v)
	}
}

func TestFindReferences(t *testing.T) {
	env := &envProvider{data: map[string]string{
		"DB_PASS": "op://vault/db/password",
		"DB_HOST": "localhost",
		"API_KEY": "op://vault/api/key",
	}}

	refs := findReferences([]conf.Provider{env})
	assert.Len(t, refs, 2)

	refMap := make(map[string]string)
	for _, r := range refs {
		refMap[r.key] = r.opRef
	}
	assert.Equal(t, "op://vault/db/password", refMap["DB_PASS"])
	assert.Equal(t, "op://vault/api/key", refMap["API_KEY"])
}

func TestNoopWhenNoToken(t *testing.T) {
	t.Setenv("OP_SERVICE_ACCOUNT_TOKEN", "")

	p := NewProvider(t.Context())
	require.NoError(t, p.Load(t.Context(), nil))

	_, ok := p.Lookup("anything")
	assert.False(t, ok)
}

func TestProviderLookupAndScan(t *testing.T) {
	// Test the data store mechanics with a manually constructed provider
	p := &provider{data: map[string]string{
		"DB_PASS": "secret123",
		"API_KEY": "key456",
	}}

	val, ok := p.Lookup("DB_PASS")
	assert.True(t, ok)
	assert.Equal(t, "secret123", val)

	_, ok = p.Lookup("MISSING")
	assert.False(t, ok)

	var scanned int
	p.Scan(func(key, value string) { scanned++ })
	assert.Equal(t, 2, scanned)
}
