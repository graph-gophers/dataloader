package dataloader

import "context"

// The Cache interface. If a custom cache is provided, it must implement this interface.
type Cache interface {
	Get(context.Context, Keyer) (Thunk, bool)
	Set(context.Context, Keyer, Thunk)
	Delete(context.Context, Keyer) bool
	Clear()
}

// NoCache implements Cache interface where all methods are noops.
// This is useful for when you don't want to cache items but still
// want to use a data loader
type NoCache struct{}

// Get is a NOOP
func (c *NoCache) Get(context.Context, Keyer) (Thunk, bool) { return nil, false }

// Set is a NOOP
func (c *NoCache) Set(context.Context, Keyer, Thunk) { return }

// Delete is a NOOP
func (c *NoCache) Delete(context.Context, Keyer) bool { return false }

// Clear is a NOOP
func (c *NoCache) Clear() { return }
