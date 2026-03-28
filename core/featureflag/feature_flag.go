// Package featureflag provides feature flag evaluation and management.
package featureflag

import "context"

// State represents the state and metadata of a feature toggle.
type State struct {
	Enabled     bool   // Whether the feature is enabled
	Description string // Optional description of the feature
	Source      string // Where the toggle state came from
}

// Service defines a client-side interface for checking feature toggle states.
type Service interface {
	// IsEnabled checks if a feature is enabled by its name.
	IsEnabled(ctx context.Context, name string) bool

	// GetState retrieves the full state and metadata for a feature toggle.
	GetState(ctx context.Context, name string) (State, error)
}

// Provider defines an interface for syncing feature toggle states from a backend.
type Provider interface {
	// Sync fetches and returns the latest feature toggle states.
	Sync(ctx context.Context) (map[string]State, error)

	// GetState retrieves the full state and metadata for a feature toggle.
	GetState(ctx context.Context, name string) (State, error)
}
