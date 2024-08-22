package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/josestg/yt-go-plugin/cache"
)

// Value represents a cache entry.
type Value struct {
	Data  string
	ExpAt time.Time
}

// Memcache is a simple in-memory cache.
type Memcache struct {
	mu    sync.RWMutex
	log   *slog.Logger
	store map[string]Value
}

// Factory is the symbol the plugin loader will try to load. It must implement the cache.Factory signature.
var Factory cache.Factory = New

// New creates a new Memcache instance.
func New(log *slog.Logger) (cache.Cache, error) {
	log.Info("[plugin/memcache] loaded")
	c := &Memcache{
		mu:    sync.RWMutex{},
		log:   log,
		store: make(map[string]Value),
	}
	return c, nil
}

func (m *Memcache) Set(ctx context.Context, key, val string, exp time.Duration) error {
	m.log.InfoContext(ctx, "[plugin/memcache] set", "key", key, "val", val, "exp", exp)
	m.mu.Lock()
	m.log.DebugContext(ctx, "[plugin/memcache] lock acquired")
	defer func() {
		m.mu.Unlock()
		m.log.DebugContext(ctx, "[plugin/memcache] lock released")
	}()

	m.store[key] = Value{
		Data:  val,
		ExpAt: time.Now().Add(exp),
	}

	return nil
}

func (m *Memcache) Get(ctx context.Context, key string) (string, error) {
	m.log.InfoContext(ctx, "[plugin/memcache] get", "key", key)
	m.mu.RLock()
	v, ok := m.store[key]
	m.mu.RUnlock()
	if !ok {
		return "", cache.ErrNotFound
	}

	if time.Now().After(v.ExpAt) {
		m.log.InfoContext(ctx, "[plugin/memcache] key expired", "key", key, "val", v)
		m.mu.Lock()
		delete(m.store, key)
		m.mu.Unlock()
		return "", cache.ErrNotFound
	}

	m.log.InfoContext(ctx, "[plugin/memcache] key found", "key", key, "val", v)
	return v.Data, nil
}
