package observer

import (
	"context"

	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

type ContextKey string

const (
	ContextKeyParentID   ContextKey = "observerParentID"
	ContextKeyGeneration ContextKey = "observerGeneration"
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
}

type Generation struct {
	ID              string
	ParentID        string
	TraceID         string
	Name            string
	Model           string
	ModelParameters types.M
	Input           []*thread.Message
	Output          *thread.Message
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

func ExtractParentIDFromContext(ctx context.Context) string {
	parentID, ok := ctx.Value(ContextKeyParentID).(string)
	if !ok {
		return ""
	}
	return parentID
}

func StoreParentIDInContext(ctx context.Context, parentID string) context.Context {
	return context.WithValue(ctx, ContextKeyParentID, parentID)
}

func ExtractGenerationFromContext(ctx context.Context) Generation {
	generation, ok := ctx.Value(ContextKeyGeneration).(Generation)
	if !ok {
		return Generation{}
	}
	return generation
}

func StoreGenerationInContext(ctx context.Context, generation Generation) context.Context {
	return context.WithValue(ctx, ContextKeyGeneration, generation)
}
