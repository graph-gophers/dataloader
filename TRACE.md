# Adding a new trace backend.

If you want to add a new tracing backend all you need to do is implement the
`Tracer` interface and pass it as an option to the dataloader on initialization.

As an example, this is how you could implement it to an OpenCensus backend.

```go
package main

import (
	"context"
	"strings"

	"github.com/graph-gophers/dataloader/v7"
	exp "go.opencensus.io/examples/exporter"
	"go.opencensus.io/trace"
)

type User struct {
	ID string
}

// OpenCensusTracer Tracer implements a tracer that can be used with the Open Tracing standard.
type OpenCensusTracer struct{}

// TraceLoad will trace a call to dataloader.LoadMany with Open Tracing
func (OpenCensusTracer) TraceLoad(ctx context.Context, key string) (context.Context, dataloader.TraceLoadFinishFunc[*User]) {
	cCtx, cSpan := trace.StartSpan(ctx, "Dataloader: load")
	cSpan.AddAttributes(
		trace.StringAttribute("dataloader.key", key),
	)
	return cCtx, func(thunk dataloader.Thunk[*User]) {
		// TODO: is there anything we should do with the results?
		cSpan.End()
	}
}

// TraceLoadMany will trace a call to dataloader.LoadMany with Open Tracing
func (OpenCensusTracer) TraceLoadMany(ctx context.Context, keys []string) (context.Context, dataloader.TraceLoadManyFinishFunc[*User]) {
	cCtx, cSpan := trace.StartSpan(ctx, "Dataloader: loadmany")
	cSpan.AddAttributes(
		trace.StringAttribute("dataloader.keys", strings.Join(keys, ",")),
	)
	return cCtx, func(thunk dataloader.ThunkMany[*User]) {
		// TODO: is there anything we should do with the results?
		cSpan.End()
	}
}

// TraceBatch will trace a call to dataloader.LoadMany with Open Tracing
func (OpenCensusTracer) TraceBatch(ctx context.Context, keys []string) (context.Context, dataloader.TraceBatchFinishFunc[*User]) {
	cCtx, cSpan := trace.StartSpan(ctx, "Dataloader: batch")
	cSpan.AddAttributes(
		trace.StringAttribute("dataloader.keys", strings.Join(keys, ",")),
	)
	return cCtx, func(results []*dataloader.Result[*User]) {
		// TODO: is there anything we should do with the results?
		cSpan.End()
	}
}

func batchFunc(ctx context.Context, keys []string) []*dataloader.Result[*User] {
	// ...loader logic goes here
}

func main() {
	//initialize an example exporter that just logs to the console
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})
	trace.RegisterExporter(&exp.PrintExporter{})
	// initialize the dataloader with your new tracer backend
	loader := dataloader.NewBatchedLoader(batchFunc, dataloader.WithTracer[string, *User](OpenCensusTracer{}))
	// initialize a context since it's not receiving one from anywhere else.
	ctx, span := trace.StartSpan(context.TODO(), "Span Name")
	defer span.End()
	// request from the dataloader as usual
	value, err := loader.Load(ctx, SomeID)()
	// ...
}
```

Don't forget to initialize the exporters of your choice and register it with `trace.RegisterExporter(&exporterInstance)`.
