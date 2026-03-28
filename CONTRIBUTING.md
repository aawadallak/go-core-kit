# Contributing to Go Core Kit

Thank you for your interest in contributing! Here's how to get started.

## Development Setup

```bash
git clone https://github.com/aawadallak/go-core-kit.git
cd go-core-kit
go build ./...
go test ./...
```

### Optional: Local Redis

```bash
make redis-local    # starts Redis on port 6379
make redis-stop     # stops and removes the container
```

## Workflow

1. **Open an issue first** for significant changes to discuss the approach
2. **Fork and branch** from `main` — use descriptive branch names (`feat/add-redis-cache`, `fix/logger-nil-panic`)
3. **Write code** following the patterns below
4. **Run checks** before pushing:
   ```bash
   go build ./...
   go test ./...
   golangci-lint run
   ```
5. **Open a PR** against `main`

## Commit Conventions

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add Redis cache TTL support
fix: handle nil pointer in logger context
refactor: simplify broker message encoding
docs: update cache package examples
chore: update go.mod dependencies
test: add idempotency manager edge cases
```

## Code Conventions

### Architecture

- **`core/`** — Interfaces only. No external dependencies beyond stdlib.
- **`plugin/`** — Concrete implementations. Import `core/` interfaces, never other plugins.
- **`pkg/`** — Shared foundation types used by both core and plugin.

### Patterns

- **Functional options** for constructors: `New(opts ...Option)`
- **Context-first** on all public methods: `func (r *Repo) Find(ctx context.Context, ...) error`
- **Generics** for type-safe abstractions where applicable
- Embed `pkg/common.Entity` for GORM-backed domain models

### Testing

- Unit tests live next to the code (`*_test.go`)
- Integration tests go in `tests/integration/` (may require Docker services)
- Use `github.com/stretchr/testify` for assertions

## Adding a New Package

### New core interface

1. Create `core/<name>/` with the interface definition
2. Keep it dependency-free (stdlib only)
3. Add the package to the table in `README.md`

### New plugin implementation

1. Create `plugin/<name>/` implementing a core interface
2. Use functional options for configuration
3. Add the package to the plugin table in `README.md`

## License

By contributing, you agree that your contributions will be licensed under the [GPL v3.0](LICENSE).
