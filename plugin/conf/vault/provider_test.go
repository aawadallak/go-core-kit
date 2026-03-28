package vault

import (
	"context"
	"fmt"
	"testing"

	"github.com/aawadallak/go-core-kit/core/conf"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcvault "github.com/testcontainers/testcontainers-go/modules/vault"
)

func setupVault(t *testing.T) *vaultapi.Client {
	t.Helper()
	ctx := t.Context()

	container, err := tcvault.Run(ctx, "hashicorp/vault:1.13.0",
		tcvault.WithToken("test-root-token"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })

	addr, err := container.HttpHostAddress(ctx)
	require.NoError(t, err)

	t.Setenv("VAULT_ADDR", addr)
	t.Setenv("VAULT_TOKEN", "test-root-token")
	t.Setenv("VAULT_MOUNT_PATH", "secret")

	client, err := vaultapi.NewClient(&vaultapi.Config{Address: addr})
	require.NoError(t, err)
	client.SetToken("test-root-token")

	return client
}

func seedSecret(t *testing.T, client *vaultapi.Client, path string, data map[string]any) {
	t.Helper()
	_, err := client.Logical().WriteWithContext(t.Context(),
		fmt.Sprintf("secret/data/%s", path),
		map[string]any{"data": data},
	)
	require.NoError(t, err)
}

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

func TestVaultProvider_ResolveSecretWithKey(t *testing.T) {
	client := setupVault(t)
	seedSecret(t, client, "myapp/db", map[string]any{
		"username": "admin",
		"password": "s3cret",
	})

	env := &envProvider{data: map[string]string{
		"DB_USER": "vault://myapp/db:username",
		"DB_PASS": "vault://myapp/db:password",
		"APP_ENV": "production",
	}}

	p := NewProvider()
	err := p.Load(t.Context(), []conf.Provider{env})
	require.NoError(t, err)

	val, ok := p.Lookup("DB_USER")
	assert.True(t, ok)
	assert.Equal(t, "admin", val)

	val, ok = p.Lookup("DB_PASS")
	assert.True(t, ok)
	assert.Equal(t, "s3cret", val)

	_, ok = p.Lookup("APP_ENV")
	assert.False(t, ok)
}

func TestVaultProvider_ResolveSingleValueSecret(t *testing.T) {
	client := setupVault(t)
	seedSecret(t, client, "myapp/api-key", map[string]any{
		"value": "key-12345",
	})

	env := &envProvider{data: map[string]string{
		"API_KEY": "vault://myapp/api-key",
	}}

	p := NewProvider()
	err := p.Load(t.Context(), []conf.Provider{env})
	require.NoError(t, err)

	val, ok := p.Lookup("API_KEY")
	assert.True(t, ok)
	assert.Equal(t, "key-12345", val)
}

func TestVaultProvider_SecretNotFound(t *testing.T) {
	_ = setupVault(t)

	env := &envProvider{data: map[string]string{
		"MISSING": "vault://nonexistent/secret",
	}}

	p := NewProvider()
	err := p.Load(t.Context(), []conf.Provider{env})
	assert.Error(t, err)
	assert.ErrorContains(t, err, "nonexistent/secret")
}

func TestVaultProvider_NoopWhenNoToken(t *testing.T) {
	t.Setenv("VAULT_TOKEN", "")

	p := NewProvider()
	err := p.Load(t.Context(), nil)
	assert.NoError(t, err)

	_, ok := p.Lookup("anything")
	assert.False(t, ok)
}
