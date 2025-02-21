package observer

import (
	"context"

	"github.com/rsest/lingoose/embedder"
	"github.com/rsest/lingoose/thread"
	"github.com/rsest/lingoose/types"
)

type ContextKey string

const (
	ContextKeyParentID         ContextKey = "observerParentID"
	ContextKeyTraceID          ContextKey = "observerTraceID"
	ContextKeyObserverInstance ContextKey = "observerInstance"
)

type Trace struct {
	ID   string
	Name string
}

type Span struct {
	ID       string
	ParentID string
	TraceID  string
	Name     string
	Input    any
	Output   any
}

type Generation struct {
	ID              string
	ParentID        string
	TraceID         string
	Name            string
	Model           string
	ModelParameters types.M
	Input           []*thread.Message
	Output          []*thread.Message
	Metadata        types.M
}

type Embedding struct {
	ID              string
	ParentID        string
	TraceID         string
	Name            string
	Model           string
	ModelParameters types.M
	Input           []string
	Output          []embedder.Embedding
	Metadata        types.M
}

type Event struct {
	ID       string
	ParentID string
	TraceID  string
	Name     string
	Metadata types.M
}

type Score struct {
	ID      string
	TraceID string
	Name    string
	Value   float64
}

func ContextValueParentID(ctx context.Context) string {
	parentID, ok := ctx.Value(ContextKeyParentID).(string)
	if !ok {
		return ""
	}
	return parentID
}

func ContextValueTraceID(ctx context.Context) string {
	traceID, ok := ctx.Value(ContextKeyTraceID).(string)
	if !ok {
		return ""
	}
	return traceID
}

func ContextValueObserverInstance(ctx context.Context) any {
	return ctx.Value(ContextKeyObserverInstance)
}

func ContextWithParentID(ctx context.Context, parentID string) context.Context {
	return context.WithValue(ctx, ContextKeyParentID, parentID)
}

func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ContextKeyTraceID, traceID)
}

func ContextWithObserverInstance(ctx context.Context, instance any) context.Context {
	return context.WithValue(ctx, ContextKeyObserverInstance, instance)
}
