package dataloader

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
)

type TraceLoadFinishFunc func(Thunk)
type TraceLoadManyFinishFunc func(ThunkMany)
type TraceBatchFinishFunc func([]*Result)

// Tracer is an interface that may be used to implement tracing.
type Tracer interface {
	// TraceLoad will trace the calls to Load
	TraceLoad(ctx context.Context, key Keyer) (context.Context, TraceLoadFinishFunc)
	// TraceLoadMany will trace the calls to LoadMany
	TraceLoadMany(ctx context.Context, keys KeyList) (context.Context, TraceLoadManyFinishFunc)
	// TraceBatch will trace data loader batches
	TraceBatch(ctx context.Context, keys KeyList) (context.Context, TraceBatchFinishFunc)
}

// OpenTracing Tracer implements a tracer that can be used with the Open Tracing standard.
type OpenTracingTracer struct{}

// TraceLoad will trace a call to dataloader.LoadMany with Open Tracing
func (OpenTracingTracer) TraceLoad(ctx context.Context, key Keyer) (context.Context, TraceLoadFinishFunc) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "Dataloader: load")

	span.SetTag("dataloader.key", key.Key())

	return spanCtx, func(thunk Thunk) {
		// TODO: is there anything we should do with the results?
		span.Finish()
	}
}

// TraceLoadMany will trace a call to dataloader.LoadMany with Open Tracing
func (OpenTracingTracer) TraceLoadMany(ctx context.Context, keys KeyList) (context.Context, TraceLoadManyFinishFunc) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "Dataloader: loadmany")

	span.SetTag("dataloader.keys", keys.Strings())

	return spanCtx, func(thunk ThunkMany) {
		// TODO: is there anything we should do with the results?
		span.Finish()
	}
}

// TraceBatch will trace a call to dataloader.LoadMany with Open Tracing
func (OpenTracingTracer) TraceBatch(ctx context.Context, keys KeyList) (context.Context, TraceBatchFinishFunc) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "Dataloader: batch")

	span.SetTag("dataloader.keys", keys.Strings())

	return spanCtx, func(results []*Result) {
		// TODO: is there anything we should do with the results?
		span.Finish()
	}
}

// NoopTracer is the default (noop) tracer
type NoopTracer struct{}

// TraceLoad is a noop function
func (NoopTracer) TraceLoad(ctx context.Context, key Keyer) (context.Context, TraceLoadFinishFunc) {
	return ctx, func(Thunk) {}
}

// TraceLoadMany is a noop function
func (NoopTracer) TraceLoadMany(ctx context.Context, keys KeyList) (context.Context, TraceLoadManyFinishFunc) {
	return ctx, func(ThunkMany) {}
}

// TraceBatch is a noop function
func (NoopTracer) TraceBatch(ctx context.Context, keys KeyList) (context.Context, TraceBatchFinishFunc) {
	return ctx, func(result []*Result) {}
}
