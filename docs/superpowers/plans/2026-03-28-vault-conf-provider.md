# Vault Configuration Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a HashiCorp Vault configuration provider to `plugin/conf/vault` that resolves `vault://` prefixed values from other providers, following the same pattern as the existing SSM provider.

**Architecture:** The Vault provider implements `conf.Provider`. During `Load()`, it scans other providers for values prefixed with `vault://`, resolves them from Vault's KV secrets engine (supporting both KV v1 and v2), and stores the plaintext values locally. Falls back to a noop provider if Vault is unavailable. Supports configurable mount path, inline mount path override (`mount//path`), and specific key selection (`path:key`).

**Tech Stack:** `github.com/hashicorp/vault/api`, `github.com/aawadallak/go-core-kit/core/conf`, testcontainers-go with Vault module for e2e tests.

---

## File Structure

```
plugin/conf/vault/
├── provider.go       -- Main vault provider: NewProvider, Load, Lookup, Scan
├── client.go         -- Vault client creation from env vars
├── parser.go         -- Path parsing (vault://, mount//, :key extraction)
├── parser_test.go    -- Unit tests for path parsing
├── noop.go           -- Noop fallback provider
└── provider_test.go  -- E2e test with Vault testcontainer
```

---

### Task 1: Path parsing utilities

**Files:**
- Create: `plugin/conf/vault/parser.go`
- Create: `plugin/conf/vault/parser_test.go`

- [ ] **Step 1: Write the failing tests for path parsing**

```go
// plugin/conf/vault/parser_test.go
package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVaultPath(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantPath   string
		wantKey    string
	}{
		{"path only", "myapp/db", "myapp/db", ""},
		{"path with key", "myapp/db:password", "myapp/db", "password"},
		{"path with colon in path", "myapp/db:user:name", "myapp/db", "user:name"},
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
		name          string
		path          string
		defaultMount  string
		wantMount     string
		wantSecret    string
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./plugin/conf/vault/... -run TestParse -v`
Expected: FAIL — files don't exist yet

- [ ] **Step 3: Implement parser.go**

```go
// plugin/conf/vault/parser.go
package vault

import "strings"

const vaultPrefix = "vault://"

type vaultRef struct {
	key       string // config key (e.g., "DB_PASS")
	vaultPath string // vault path without prefix (e.g., "myapp/db:password")
}

type kv struct {
	key   string
	value string
}

// parseVaultPath splits "path/to/secret:key" into secret path and optional key.
func parseVaultPath(raw string) (secretPath, secretKey string) {
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return raw, ""
}

// extractMountPath extracts inline mount override from "mount//path" format.
// Falls back to defaultMount if no "//" delimiter is present.
func extractMountPath(path, defaultMount string) (mount, secretPath string) {
	parts := strings.SplitN(path, "//", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return defaultMount, path
}

// extractVaultReferences scans key-value pairs for vault:// prefixed values.
func extractVaultReferences(pairs []kv) []vaultRef {
	var refs []vaultRef
	for _, p := range pairs {
		if strings.HasPrefix(p.value, vaultPrefix) {
			refs = append(refs, vaultRef{
				key:       p.key,
				vaultPath: strings.TrimPrefix(p.value, vaultPrefix),
			})
		}
	}
	return refs
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./plugin/conf/vault/... -run TestParse -v && go test ./plugin/conf/vault/... -run TestExtract -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add plugin/conf/vault/parser.go plugin/conf/vault/parser_test.go
git commit -m "feat(conf): add vault path parsing utilities"
```

---

### Task 2: Vault client and noop fallback

**Files:**
- Create: `plugin/conf/vault/client.go`
- Create: `plugin/conf/vault/noop.go`

- [ ] **Step 1: Create client.go**

```go
// plugin/conf/vault/client.go
package vault

import (
	"fmt"
	"os"

	vaultapi "github.com/hashicorp/vault/api"
)

// newVaultClient creates a Vault API client from environment variables.
// Required: VAULT_TOKEN (or returns error).
// Optional: VAULT_ADDR (default: http://127.0.0.1:8200), VAULT_MOUNT_PATH (default: secret).
func newVaultClient() (*vaultapi.Client, string, error) {
	addr := os.Getenv("VAULT_ADDR")
	if addr == "" {
		addr = "http://127.0.0.1:8200"
	}

	mountPath := os.Getenv("VAULT_MOUNT_PATH")
	if mountPath == "" {
		mountPath = "secret"
	}

	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		return nil, "", fmt.Errorf("VAULT_TOKEN environment variable is not set")
	}

	client, err := vaultapi.NewClient(&vaultapi.Config{Address: addr})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create vault client: %w", err)
	}

	client.SetToken(token)

	return client, mountPath, nil
}
```

- [ ] **Step 2: Create noop.go**

```go
// plugin/conf/vault/noop.go
package vault

import (
	"context"

	"github.com/aawadallak/go-core-kit/core/conf"
)

type noopProvider struct{}

var _ conf.Provider = (*noopProvider)(nil)

func (n *noopProvider) Load(_ context.Context, _ []conf.Provider) error { return nil }
func (n *noopProvider) Lookup(_ string) (string, bool)                  { return "", false }
func (n *noopProvider) Scan(_ conf.ScanFunc)                            {}
```

- [ ] **Step 3: Add vault/api dependency**

Run: `go get github.com/hashicorp/vault/api`

- [ ] **Step 4: Verify compilation**

Run: `go build ./plugin/conf/vault/...`

- [ ] **Step 5: Commit**

```bash
git add plugin/conf/vault/client.go plugin/conf/vault/noop.go go.mod go.sum
git commit -m "feat(conf): add vault client and noop fallback"
```

---

### Task 3: Main vault provider

**Files:**
- Create: `plugin/conf/vault/provider.go`

- [ ] **Step 1: Create provider.go**

```go
// Package vault provides a HashiCorp Vault configuration provider for core/conf.
package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/aawadallak/go-core-kit/core/conf"
	vaultapi "github.com/hashicorp/vault/api"
)

type provider struct {
	data      map[string]string
	client    *vaultapi.Client
	mountPath string
}

var _ conf.Provider = (*provider)(nil)

// NewProvider creates a Vault configuration provider.
// Returns a noop provider if Vault is unavailable (missing token, connection failure).
func NewProvider() conf.Provider {
	client, mountPath, err := newVaultClient()
	if err != nil {
		return &noopProvider{}
	}
	return &provider{
		client:    client,
		mountPath: mountPath,
		data:      make(map[string]string),
	}
}

func (p *provider) Lookup(key string) (string, bool) {
	v, ok := p.data[key]
	return v, ok
}

func (p *provider) Scan(fn conf.ScanFunc) {
	for k, v := range p.data {
		fn(k, v)
	}
}

// Load scans other providers for vault:// prefixed values and resolves them.
func (p *provider) Load(ctx context.Context, others []conf.Provider) error {
	refs := p.findReferences(others)

	for _, ref := range refs {
		secretPath, secretKey := parseVaultPath(ref.vaultPath)
		mountPath, actualPath := extractMountPath(secretPath, p.mountPath)

		data, err := p.loadSecret(ctx, actualPath, mountPath)
		if err != nil {
			return fmt.Errorf("failed to load vault secret %s: %w", secretPath, err)
		}

		value, err := extractValue(data, secretPath, secretKey)
		if err != nil {
			return err
		}

		p.data[ref.key] = value
	}

	return nil
}

func (p *provider) findReferences(others []conf.Provider) []vaultRef {
	var pairs []kv
	for _, other := range others {
		other.Scan(func(key, value string) {
			pairs = append(pairs, kv{key: key, value: value})
		})
	}
	return extractVaultReferences(pairs)
}

// loadSecret reads a secret from Vault, trying KV v2 first then KV v1.
func (p *provider) loadSecret(ctx context.Context, secretPath, mountPath string) (map[string]any, error) {
	path := strings.TrimPrefix(secretPath, mountPath+"/")

	// Try KV v2: mount/data/path
	kv2Path := fmt.Sprintf("%s/data/%s", mountPath, path)
	secret, err := p.client.Logical().ReadWithContext(ctx, kv2Path)
	if err == nil && secret != nil && secret.Data != nil {
		if data, ok := secret.Data["data"].(map[string]any); ok && len(data) > 0 {
			return data, nil
		}
	}

	// Try KV v1: mount/path
	kv1Path := fmt.Sprintf("%s/%s", mountPath, path)
	secret, err = p.client.Logical().ReadWithContext(ctx, kv1Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read vault secret: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found: %s", secretPath)
	}

	// Filter vault metadata for KV v1
	result := make(map[string]any)
	for k, v := range secret.Data {
		if k != "lease_id" && k != "lease_duration" && k != "renewable" && k != "metadata" {
			result[k] = v
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("secret found but contains no data: %s", secretPath)
	}

	return result, nil
}

// extractValue picks the right value from secret data based on the key selector.
func extractValue(data map[string]any, secretPath, secretKey string) (string, error) {
	if secretKey != "" {
		val, ok := data[secretKey]
		if !ok {
			return "", fmt.Errorf("key %q not found in secret %s", secretKey, secretPath)
		}
		str, ok := val.(string)
		if !ok {
			return "", fmt.Errorf("key %q in secret %s is not a string", secretKey, secretPath)
		}
		return str, nil
	}

	if len(data) == 1 {
		for _, v := range data {
			if str, ok := v.(string); ok {
				return str, nil
			}
		}
	}

	return "", fmt.Errorf("secret %s has multiple keys, specify which one (e.g., vault://%s:keyname)", secretPath, secretPath)
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./plugin/conf/vault/...`

- [ ] **Step 3: Commit**

```bash
git add plugin/conf/vault/provider.go
git commit -m "feat(conf): add vault configuration provider with KV v1/v2 support"
```

---

### Task 4: E2e test with Vault testcontainer

**Files:**
- Create: `plugin/conf/vault/provider_test.go`

- [ ] **Step 1: Write e2e test**

```go
// plugin/conf/vault/provider_test.go
package vault

import (
	"context"
	"fmt"
	"testing"

	"github.com/aawadallak/go-core-kit/core/conf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcvault "github.com/testcontainers/testcontainers-go/modules/vault"
	vaultapi "github.com/hashicorp/vault/api"
)

func setupVault(t *testing.T) *vaultapi.Client {
	t.Helper()
	ctx := t.Context()

	container, err := tcvault.Run(ctx, "hashicorp/vault:1.13.0",
		tcvault.WithToken("test-root-token"),
		tcvault.WithInitCommand("secrets enable -version=2 -path=secret kv"),
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

// envProvider simulates upstream config with vault:// references.
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

	// Non-vault keys should not be in vault provider
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
	assert.Contains(t, err.Error(), "secret not found")
}

func TestVaultProvider_NoopWhenNoToken(t *testing.T) {
	t.Setenv("VAULT_TOKEN", "")

	p := NewProvider()
	err := p.Load(t.Context(), nil)
	assert.NoError(t, err) // noop should not error

	_, ok := p.Lookup("anything")
	assert.False(t, ok)
}
```

- [ ] **Step 2: Add testcontainers vault module dependency**

Run: `go get github.com/testcontainers/testcontainers-go/modules/vault`

- [ ] **Step 3: Run e2e tests**

Run: `go test ./plugin/conf/vault/... -v -timeout 120s`
Expected: PASS (requires Docker)

- [ ] **Step 4: Commit**

```bash
git add plugin/conf/vault/provider_test.go go.mod go.sum
git commit -m "test(conf): add vault provider e2e tests with testcontainers"
```

---

### Task 5: Update docs

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`

- [ ] **Step 1: Add vault to README plugin table**

In the Plugin Implementations table, add after the `plugin/conf/ssm` row:

```markdown
| `plugin/conf/vault` | [HashiCorp Vault](https://www.vaultproject.io) | `core/conf` |
```

- [ ] **Step 2: Add to CHANGELOG under [Unreleased]**

Add under the existing Added section:

```markdown
- **`plugin/conf/vault`** — HashiCorp Vault configuration provider with KV v1/v2 support, `vault://` prefix resolution, inline mount path override, and graceful noop fallback.
```

- [ ] **Step 3: Commit**

```bash
git add README.md CHANGELOG.md
git commit -m "docs: add vault provider to README and CHANGELOG"
```
