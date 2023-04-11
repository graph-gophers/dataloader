package dataloader

import "context"

// DetachedContext returns a new context detached from the lifetime
// of ctx, but which still returns the values of ctx.
//
// DetachedContext can be used to maintain the trace context required
// to correlate events, but where the operation is "fire-and-forget",
// and should not be affected by the deadline or cancellation of ctx.
func DetachedContext(ctx context.Context) context.Context {
	return &detachedContext{Context: context.Background(), orig: ctx}
}

type detachedContext struct {
	context.Context
	orig context.Context
}

// Value returns c.orig.Value(key).
func (c *detachedContext) Value(key interface{}) interface{} {
	return c.orig.Value(key)
}
