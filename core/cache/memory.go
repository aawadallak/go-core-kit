package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type inMemoryCache struct {
	data     map[string]Item
	mu       sync.RWMutex
	isClosed bool
}

func (i *inMemoryCache) Set(ctx context.Context, item Item) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.isClosed {
		return ErrClosed
	}

	i.data[item.Key] = item

	if item.ExpiresIn > 0 {
		time.AfterFunc(item.ExpiresIn, func() {
			//nolint:errcheck
			i.Delete(context.TODO(), item.Key)
		})
	}

	return nil
}

func (i *inMemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	item, ok := i.data[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	data, ok := item.Value.([]byte)
	if ok {
		return data, nil
	}

	data, err := json.Marshal(item.Value)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (i *inMemoryCache) Delete(ctx context.Context, key string) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	delete(i.data, key)
	return nil
}

func (i *inMemoryCache) Close(ctx context.Context) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.isClosed = true
	return nil
}

// NewInMemoryCache creates and returns a new in-memory cache provider.
// The in-memory cache provider stores items in memory and is useful for local caching.
// Returns a new in-memory cache provider.
func NewInMemoryCache() Provider {
	return &inMemoryCache{
		data: make(map[string]Item),
	}
}
