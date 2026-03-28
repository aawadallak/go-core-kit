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

func NewProvider() conf.Provider {
	client, mountPath, err := newVaultClient()
	if err != nil {
		return &noopProvider{}
	}
	return &provider{client: client, mountPath: mountPath, data: make(map[string]string)}
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
