package vault

import "strings"

const vaultPrefix = "vault://"

type vaultRef struct {
	key       string
	vaultPath string
}

type kv struct {
	key   string
	value string
}

func parseVaultPath(raw string) (secretPath, secretKey string) {
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return raw, ""
}

func extractMountPath(path, defaultMount string) (mount, secretPath string) {
	parts := strings.SplitN(path, "//", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return defaultMount, path
}

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
