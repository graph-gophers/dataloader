package otel_test

import (
	"testing"

	"github.com/graph-gophers/dataloader/v8"
	"github.com/graph-gophers/dataloader/v8/trace/otel"
)

func TestInterfaceImplementation(t *testing.T) {
	type User struct {
		ID        uint
		FirstName string
		LastName  string
		Email     string
	}
	var _ dataloader.Tracer[string, int] = otel.Tracer[string, int]{}
	var _ dataloader.Tracer[string, string] = otel.Tracer[string, string]{}
	var _ dataloader.Tracer[uint, User] = otel.Tracer[uint, User]{}
	// check compatibility with loader options
	dataloader.WithTracer[uint, User](&otel.Tracer[uint, User]{})
}
