package dataloader

import (
	"context"
)

type Key[K comparable] interface {
	Raw() K
	Context() context.Context
}

type Keys[K comparable] []Key[K]

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

func ContextKey[K comparable](ctx context.Context, key K) Key[K] {
	return &ckey[K]{root: key, ctx: ctx}
}

func ContextKeys[K comparable](ctx context.Context, keys []K) Keys[K] {
	result := make(Keys[K], len(keys))
	for i := range keys {
		result[i] = ContextKey(ctx, keys[i])
	}

	return result
}

func (k Keys[K]) Raw() []K {
	result := make([]K, len(k))
	for i := range k {
		result[i] = k[i].Raw()
	}

	return result
}
