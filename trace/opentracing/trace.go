package opentracing

import (
	"context"

	"github.com/graph-gophers/dataloader/v8"

	"github.com/opentracing/opentracing-go"
)

// Tracer implements a tracer that can be used with the Open Tracing standard.
type Tracer[K any, V any] struct{}

// TraceLoad will trace a call to dataloader.LoadMany with Open Tracing.
func (Tracer[K, V]) TraceLoad(ctx context.Context, key dataloader.Key[K]) (context.Context, dataloader.TraceLoadFinishFunc[V]) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "Dataloader: load")

	span.SetTag("dataloader.key", key.String())

	return spanCtx, func(thunk dataloader.Thunk[V]) {
		span.Finish()
	}
}

// TraceLoadMany will trace a call to dataloader.LoadMany with Open Tracing.
func (Tracer[K, V]) TraceLoadMany(ctx context.Context, keys dataloader.Keys[K]) (context.Context, dataloader.TraceLoadManyFinishFunc[V]) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "Dataloader: loadmany")

	span.SetTag("dataloader.keys", keys.Keys())

	return spanCtx, func(thunk dataloader.ThunkMany[V]) {
		span.Finish()
	}
}

// TraceBatch will trace a call to dataloader.LoadMany with Open Tracing.
func (Tracer[K, V]) TraceBatch(ctx context.Context, keys dataloader.Keys[K]) (context.Context, dataloader.TraceBatchFinishFunc[V]) {
	span, spanCtx := opentracing.StartSpanFromContext(ctx, "Dataloader: batch")

	span.SetTag("dataloader.keys", keys.Keys())

	return spanCtx, func(results []*dataloader.Result[V]) {
		span.Finish()
	}
}
