package otel

import (
	"context"

	"github.com/graph-gophers/dataloader/v8"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Tracer implements a tracer that can be used with the Open Tracing standard.
type Tracer[K any, V any] struct {
	tr trace.Tracer
}

func NewTracer[K any, V any](tr trace.Tracer) *Tracer[K, V] {
	return &Tracer[K, V]{tr: tr}
}

func (t *Tracer[K, V]) Tracer() trace.Tracer {
	if t.tr != nil {
		return t.tr
	}
	return otel.Tracer("graph-gophers/dataloader")
}

// TraceLoad will trace a call to dataloader.LoadMany with Open Tracing.
func (t Tracer[K, V]) TraceLoad(ctx context.Context, key dataloader.Key[K]) (context.Context, dataloader.TraceLoadFinishFunc[V]) {
	spanCtx, span := t.Tracer().Start(ctx, "Dataloader: load")

	span.SetAttributes(attribute.String("dataloader.key", key.String()))

	return spanCtx, func(thunk dataloader.Thunk[V]) {
		span.End()
	}
}

// TraceLoadMany will trace a call to dataloader.LoadMany with Open Tracing.
func (t Tracer[K, V]) TraceLoadMany(ctx context.Context, keys dataloader.Keys[K]) (context.Context, dataloader.TraceLoadManyFinishFunc[V]) {
	spanCtx, span := t.Tracer().Start(ctx, "Dataloader: loadmany")

	span.SetAttributes(attribute.StringSlice("dataloader.keys", keys.Keys()))

	return spanCtx, func(thunk dataloader.ThunkMany[V]) {
		span.End()
	}
}

// TraceBatch will trace a call to dataloader.LoadMany with Open Tracing.
func (t Tracer[K, V]) TraceBatch(ctx context.Context, keys dataloader.Keys[K]) (context.Context, dataloader.TraceBatchFinishFunc[V]) {
	spanCtx, span := t.Tracer().Start(ctx, "Dataloader: batch")

	span.SetAttributes(attribute.StringSlice("dataloader.keys", keys.Keys()))

	return spanCtx, func(results []*dataloader.Result[V]) {
		span.End()
	}
}
