package featureflag

import (
	"context"
	"maps"
	"sync"
	"time"

	"github.com/aawadallak/go-core-kit/core/logger"
)

// FeatureFlagService is an implementation of the Client interface that uses a Provider.
type FeatureFlagService struct {
	provider       Provider
	toggles        map[string]State
	mu             sync.RWMutex
	isCacheEnabled bool
}

// Option defines a functional option for configuring the ToggleService.
type Option func(*FeatureFlagService)

// WithAutoSync configures the ToggleService to automatically sync with the provider at the given interval.
func WithAutoSync(ctx context.Context, interval time.Duration) Option {
	return func(s *FeatureFlagService) {
		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := s.Sync(ctx); err != nil {
						logger.Of(ctx).ErrorS(
							"FeatureFlagService:Sync:Error",
							logger.WithValue("error", err.Error()),
						)
					}
				}
			}
		}()
	}
}

func WithCacheEnabled(enabled bool) Option {
	return func(s *FeatureFlagService) {
		s.isCacheEnabled = enabled
	}
}

// NewToggleService creates a new ToggleService with the given provider and defaults.
func NewToggleService(provider Provider, opts ...Option) (*FeatureFlagService, error) {
	s := &FeatureFlagService{
		provider: provider,
		toggles:  make(map[string]State),
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Perform initial sync
	if err := s.Sync(context.Background()); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *FeatureFlagService) IsEnabled(ctx context.Context, name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.isCacheEnabled {
		if state, exists := s.toggles[name]; exists {
			return state.Enabled
		}
	}

	state, err := s.provider.GetState(ctx, name)
	if err != nil {
		return false
	}

	return state.Enabled
}

func (s *FeatureFlagService) GetState(ctx context.Context, name string) (State, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.isCacheEnabled {
		if state, exists := s.toggles[name]; exists {
			return state, nil
		}
	}

	state, err := s.provider.GetState(ctx, name)
	if err != nil {
		return State{Source: "default"}, nil
	}

	return state, nil
}

// Sync updates the local toggle states using the provider.
func (s *FeatureFlagService) Sync(ctx context.Context) error {
	toggles, err := s.provider.Sync(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	maps.Copy(s.toggles, toggles)

	return nil
}
