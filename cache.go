package dataloader

import "context"

// The Cache interface. If a custom cache is provided, it must implement this interface.
type Cache[K any, V any] interface {
	Get(context.Context, Key[K]) (Thunk[V], bool)
	Set(context.Context, Key[K], Thunk[V])
	Delete(context.Context, Key[K]) bool
	Clear()
}

var _ Cache[any, any] = (*NoCache[any, any])(nil)

// NoCache implements Cache interface where all methods are noops.
// This is useful for when you don't want to cache items but still
// want to use a data loader
type NoCache[K any, V any] struct{}

// Get is a NOOP
func (c *NoCache[K, V]) Get(context.Context, Key[K]) (Thunk[V], bool) { return nil, false }

// Set is a NOOP
func (c *NoCache[K, V]) Set(context.Context, Key[K], Thunk[V]) { return }

// Delete is a NOOP
func (c *NoCache[K, V]) Delete(context.Context, Key[K]) bool { return false }

// Clear is a NOOP
func (c *NoCache[K, V]) Clear() { return }
