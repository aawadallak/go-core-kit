package vault

import (
	"errors"
	"fmt"
	"os"

	vaultapi "github.com/hashicorp/vault/api"
)

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
		return nil, "", errors.New("VAULT_TOKEN environment variable is not set")
	}
	client, err := vaultapi.NewClient(&vaultapi.Config{Address: addr})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create vault client: %w", err)
	}
	client.SetToken(token)
	return client, mountPath, nil
}
