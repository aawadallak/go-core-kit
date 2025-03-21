package cache

import "sync"

var global Cache
var once sync.Once

// Instance returns the global Cache instance.
// If the global instance does not exist, it is created using a new in-memory cache.
// Returns the global Cache instance.
func Instance() Cache {
	once.Do(func() {
		if global == nil {
			global = NewCache(NewInMemoryCache())
		}
	})

	return global
}

// SetGlobal sets the global Cache instance to the specified Cache instance.
// If the specified Cache instance is nil, SetGlobal does nothing.
func SetGlobal(cache Cache) {
	if cache == nil {
		return
	}

	global = cache
}
