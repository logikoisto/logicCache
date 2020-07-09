package logicCache

import (
	"context"
	"sync"
	"time"
)

// TODO: add Config

type CompareFn func(a, b interface{}) interface{}
type Cache struct {
	expireFn   ExpireFn
	defaultTTL time.Duration
	m          mapType

	// thread safe:
	mu       sync.RWMutex
	shutdown <-chan struct{} // context cancel
	wg       sync.WaitGroup

	// thread unsafe(TODO):
	CompareFn CompareFn

	// TODO: add reset TTL on hit/set
	// resetTTL bool
}

type mapType map[string]item

// New creates fully functional cache.
//
// ctx is used for shutdown.
func New(ctx context.Context, defaultTTL time.Duration, expireFn ExpireFn) *Cache {
	if expireFn == nil {
		panic("expireFn can't be nil")
	}

	return &Cache{
		m:          make(mapType),
		shutdown:   ctx.Done(),
		defaultTTL: defaultTTL,
		expireFn:   expireFn,
	}
}

// SetWithTTL sets the value for a key with default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL sets the value for a key.
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	i, found := c.m[key]
	if found {
		// update
		if c.CompareFn != nil {
			value = c.CompareFn(i.get(), value)
		}

		i.set(value)
		c.m[key] = i

		return
	}

	// create new one
	c.wg.Add(1)
	i = newItem(value)
	c.m[key] = i
	go c.newJanitor(key, i.done, ttl)
}

// Get returns the value stored in the map for a key, or nil if no
// value is present.
// The found result indicates whether value was found in the cache.
func (c *Cache) Get(key string) (value interface{}, found bool) {
	c.mu.RLock()
	i, found := c.m[key]
	c.mu.RUnlock()

	if !found {
		return nil, false
	}

	return i.get(), true
}

// Delete deletes the value for a key.
func (c *Cache) Delete(key string) {
	i, found := c.delete(key)
	if !found {
		// nothing to do
		return
	}

	// janitor can be closed in the background as item already deleted from the map:
	i.delete()
}

// Done returns a channel that will be closed when work done(ExpireFn called for all records).
// This channel would not be closed if there are no records and context is not canceled(expired, etc.).
func (c *Cache) Done() <-chan struct{} {
	done := make(chan struct{})

	go func() {
		<-c.shutdown
		// after context is canceled, wait for all janitors to be closed:
		c.wg.Wait()
		close(done)
	}()

	return done
}

// delete - doesn't delete vault
func (c *Cache) delete(key string) (item, bool) {
	c.mu.Lock()

	i, found := c.m[key]
	if !found {
		// nothing to do
		c.mu.Unlock()
		return i, false
	}

	delete(c.m, key)
	c.mu.Unlock()

	c.expireFn(key, i.get())

	return i, true
}

func (c *Cache) newJanitor(key string, done <-chan struct{}, ttl time.Duration) {
	defer c.wg.Done()

	timer := time.NewTimer(ttl)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			c.delete(key)
			return

		case <-c.shutdown:
			// global shutdown
			c.delete(key)
			return

		case <-done:
			// item deleted and ExpireFn already called
			return
		}
	}
}
