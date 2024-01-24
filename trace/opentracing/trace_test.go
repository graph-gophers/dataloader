package opentracing_test

import (
	"testing"

	"github.com/graph-gophers/dataloader/v8"
	"github.com/graph-gophers/dataloader/v8/trace/opentracing"
)

func TestInterfaceImplementation(t *testing.T) {
	type User struct {
		ID        uint
		FirstName string
		LastName  string
		Email     string
	}
	var _ dataloader.Tracer[string, int] = opentracing.Tracer[string, int]{}
	var _ dataloader.Tracer[string, string] = opentracing.Tracer[string, string]{}
	var _ dataloader.Tracer[uint, User] = opentracing.Tracer[uint, User]{}
	// check compatibility with loader options
	dataloader.WithTracer[uint, User](&opentracing.Tracer[uint, User]{})
}
