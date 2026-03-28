package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVaultPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantPath string
		wantKey  string
	}{
		{"path only", "myapp/db", "myapp/db", ""},
		{"path with key", "myapp/db:password", "myapp/db", "password"},
		{"path with colon in key", "myapp/db:user:name", "myapp/db", "user:name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, key := parseVaultPath(tt.input)
			assert.Equal(t, tt.wantPath, path)
			assert.Equal(t, tt.wantKey, key)
		})
	}
}

func TestExtractMountPath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		defaultMount string
		wantMount    string
		wantSecret   string
	}{
		{"default mount", "myapp/db", "secret", "secret", "myapp/db"},
		{"inline mount", "kv//myapp/db", "secret", "kv", "myapp/db"},
		{"no double slash", "kv/myapp/db", "secret", "secret", "kv/myapp/db"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mount, secret := extractMountPath(tt.path, tt.defaultMount)
			assert.Equal(t, tt.wantMount, mount)
			assert.Equal(t, tt.wantSecret, secret)
		})
	}
}

func TestExtractVaultReferences(t *testing.T) {
	refs := extractVaultReferences([]kv{
		{"DB_PASS", "vault://myapp/db:password"},
		{"APP_NAME", "my-service"},
		{"API_KEY", "vault://myapp/api"},
	})
	assert.Len(t, refs, 2)
	assert.Equal(t, "DB_PASS", refs[0].key)
	assert.Equal(t, "myapp/db:password", refs[0].vaultPath)
	assert.Equal(t, "API_KEY", refs[1].key)
	assert.Equal(t, "myapp/api", refs[1].vaultPath)
}
