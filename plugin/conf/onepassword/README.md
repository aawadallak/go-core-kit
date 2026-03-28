# 1Password Configuration Provider

A `core/conf.Provider` implementation that resolves secrets from [1Password](https://1password.com) using the official SDK and service account authentication.

## Prerequisites

- [1Password CLI](https://developer.1password.com/docs/cli/get-started/) (`op`) installed
- A 1Password account with access to create service accounts

## Setup

### 1. Create a vault

```bash
op vault create "my-app-secrets"
```

### 2. Add secrets to the vault

```bash
# Create an item with multiple fields
op item create --vault "my-app-secrets" \
  --category login \
  --title "database" \
  username=admin \
  password=s3cret \
  host=db.example.com

# Create a single-field secret
op item create --vault "my-app-secrets" \
  --category "API Credential" \
  --title "stripe" \
  credential=sk_live_abc123
```

### 3. Create a service account

```bash
op service-account create "my-app-sa" \
  --vault "my-app-secrets:read_items"
```

This outputs a service account token. Store it securely — you'll need it as `OP_SERVICE_ACCOUNT_TOKEN`.

### 4. Set the environment variable

```bash
export OP_SERVICE_ACCOUNT_TOKEN="ops_..."
```

## Usage

Reference secrets in your environment or config using the `op://` prefix:

```
op://vault-name/item-name/field-name
```

### Example

```bash
# Environment variables with 1Password references
export DB_USER="op://my-app-secrets/database/username"
export DB_PASS="op://my-app-secrets/database/password"
export DB_HOST="op://my-app-secrets/database/host"
export STRIPE_KEY="op://my-app-secrets/stripe/credential"
export APP_ENV="production"  # plain values are left as-is
```

```go
package main

import (
    "context"

    "github.com/aawadallak/go-core-kit/core/conf"
    op "github.com/aawadallak/go-core-kit/plugin/conf/onepassword"
)

func main() {
    ctx := context.Background()

    cfg := conf.New(
        conf.WithProvider(op.NewProvider(ctx)),
    )

    // Resolved from 1Password
    dbPass := cfg.GetString("DB_PASS")     // "s3cret"
    stripeKey := cfg.GetString("STRIPE_KEY") // "sk_live_abc123"

    // Plain value, untouched
    appEnv := cfg.GetString("APP_ENV")     // "production"
}
```

## How It Works

1. During `Load()`, the provider scans all upstream providers (e.g., environment variables) for values prefixed with `op://`
2. Each `op://vault/item/field` reference is resolved via the 1Password SDK's `Secrets().Resolve()` method
3. The resolved plaintext value replaces the reference in the config
4. Non-`op://` values are ignored and left for other providers to handle

## Configuration

| Environment Variable | Required | Default | Description |
|---|---|---|---|
| `OP_SERVICE_ACCOUNT_TOKEN` | Yes | — | 1Password service account token |

If `OP_SERVICE_ACCOUNT_TOKEN` is not set, the provider silently falls back to a noop (returns no values, no errors). This allows the same code to run in environments without 1Password access.

## Secret Reference Format

```
op://vault-name/item-name/field-name
```

| Segment | Description |
|---------|-------------|
| `vault-name` | Name of the 1Password vault |
| `item-name` | Title of the item in the vault |
| `field-name` | Specific field within the item |

## Useful CLI Commands

```bash
# List vaults
op vault list

# List items in a vault
op item list --vault "my-app-secrets"

# Get a specific item
op item get "database" --vault "my-app-secrets"

# Read a specific field
op read "op://my-app-secrets/database/password"

# Delete an item
op item delete "database" --vault "my-app-secrets"

# Delete a vault
op vault delete "my-app-secrets"
```
