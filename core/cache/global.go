package cache

import "sync"

// cacheInstance manages the global Cache instance with thread-safe access.
type cacheInstance struct {
	cache Cache
	mu    sync.Mutex // For write protection
	once  sync.Once  // For lazy initialization
}

var global = &cacheInstance{}

// Instance returns the global Cache instance.
// If the instance hasn't been initialized, it lazily creates a default in-memory cache.
// The returned Cache is guaranteed to be non-nil.
//
// Returns:
//
//	Cache - The current global cache instance
func Instance() Cache {
	if global.cache == nil {
		global.once.Do(func() {
			global.cache = New(NewInMemoryCache())
		})
	}

	return global.cache
}

// SetInstance sets the global Cache instance to the specified Cache.
// It is thread-safe and does nothing if the provided cache is nil.
// This can be called before or after Instance() to override the default.
//
// Parameters:
//
//	cache - The new Cache instance to set as global
func SetInstance(cache Cache) {
	if cache == nil {
		return
	}

	global.mu.Lock()
	defer global.mu.Unlock()

	global.cache = cache
}
