package dataloader

import (
	"context"
)

var _ Key[string] = &ckey[string]{}

// Key interface for save real context if it needed
type Key[K comparable] interface {
	Raw() K
	Context() context.Context
}

type ckey[K comparable] struct {
	root K
	ctx  context.Context
}

func (c *ckey[K]) Raw() K {
	return c.root
}

func (c *ckey[K]) Context() context.Context {
	return c.ctx
}

// ContextKey make comparable Key with save real context
func ContextKey[K comparable](ctx context.Context, key K) Key[K] {
	return &ckey[K]{root: key, ctx: ctx}
}

type Keys[K comparable] []Key[K]

func (k Keys[K]) Raw() []K {
	res := make([]K, len(k))
	for i := range k {
		res[i] = k[i].Raw()
	}
	return res
}
