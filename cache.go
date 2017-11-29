package dataloader

import "context"

// The Cache interface. If a custom cache is provided, it must implement this interface.
type Cache interface {
	Get(context.Context, string) (Thunk, bool)
	Set(context.Context, string, Thunk)
	Delete(context.Context, string) bool
	Clear()
}

// NoCache implements Cache interface where all methods are noops.
// This is useful for when you don't want to cache items but still
// want to use a data loader
type NoCache struct{}

// Get is a NOOP
func (c *NoCache) Get(context.Context, string) (Thunk, bool) { return nil, false }

// Set is a NOOP
func (c *NoCache) Set(context.Context, string, Thunk) { return }

// Delete is a NOOP
func (c *NoCache) Delete(context.Context, string) bool { return false }

// Clear is a NOOP
func (c *NoCache) Clear() { return }
