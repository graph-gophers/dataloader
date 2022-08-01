package dataloader

import (
	"context"
)

type TraceLoadFinishFunc[V any] func(Thunk[V])
type TraceLoadManyFinishFunc[V any] func(ThunkMany[V])
type TraceBatchFinishFunc[V any] func([]*Result[V])

// Tracer is an interface that may be used to implement tracing.
type Tracer[K any, V any] interface {
	// TraceLoad will trace the calls to Load.
	TraceLoad(ctx context.Context, key Key[K]) (context.Context, TraceLoadFinishFunc[V])
	// TraceLoadMany will trace the calls to LoadMany.
	TraceLoadMany(ctx context.Context, keys Keys[K]) (context.Context, TraceLoadManyFinishFunc[V])
	// TraceBatch will trace data loader batches.
	TraceBatch(ctx context.Context, keys Keys[K]) (context.Context, TraceBatchFinishFunc[V])
}

// NoopTracer is the default (noop) tracer
type NoopTracer[K any, V any] struct{}

// TraceLoad is a noop function
func (NoopTracer[K, V]) TraceLoad(ctx context.Context, _ Key[K]) (context.Context, TraceLoadFinishFunc[V]) {
	return ctx, func(Thunk[V]) {}
}

// TraceLoadMany is a noop function
func (NoopTracer[K, V]) TraceLoadMany(ctx context.Context, _ Keys[K]) (context.Context, TraceLoadManyFinishFunc[V]) {
	return ctx, func(ThunkMany[V]) {}
}

// TraceBatch is a noop function
func (NoopTracer[K, V]) TraceBatch(ctx context.Context, _ Keys[K]) (context.Context, TraceBatchFinishFunc[V]) {
	return ctx, func(result []*Result[V]) {}
}
