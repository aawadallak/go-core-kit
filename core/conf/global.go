package conf

import (
	"context"
	"sync"
)

// valueMapInstance manages the global ValueMap instance with thread-safe access.
type valueMapInstance struct {
	valueMap ValueMap
	mu       sync.Mutex // protects valueMap during writes
	once     sync.Once  // ensures single initialization
}

var global = &valueMapInstance{}

// Instance returns the global ValueMap instance.
// If not yet initialized, it creates a default in-memory implementation.
// The returned ValueMap is always non-nil.
//
// This function is thread-safe and uses lazy initialization.
func Instance() ValueMap {
	if global.valueMap == nil {
		global.once.Do(func() {
			global.valueMap = New(context.TODO(), WithMustLoad())
		})
	}
	return global.valueMap
}

// SetInstance replaces the global ValueMap instance with the provided one.
// If valueMap is nil, this function does nothing.
// This function is thread-safe and can be called before or after Instance()
// to override the default implementation.
func SetInstance(valueMap ValueMap) {
	if valueMap == nil {
		return
	}
	global.mu.Lock()
	defer global.mu.Unlock()
	global.valueMap = valueMap
}
