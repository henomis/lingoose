package main

import (
	"context"

	"github.com/rsest/lingoose/llm/openai"
	"github.com/rsest/lingoose/observer"
	"github.com/rsest/lingoose/observer/langfuse"
	"github.com/rsest/lingoose/thread"
)

func main() {
	ctx := context.Background()

	o := langfuse.New(ctx)
	trace, err := o.Trace(&observer.Trace{Name: "Who are you"})
	if err != nil {
		panic(err)
	}

	ctx = observer.ContextWithObserverInstance(ctx, o)
	ctx = observer.ContextWithTraceID(ctx, trace.ID)

	span, err := o.Span(
		&observer.Span{
			TraceID: trace.ID,
			Name:    "SPAN",
		},
	)
	if err != nil {
		panic(err)
	}

	ctx = observer.ContextWithParentID(ctx, span.ID)

	openaillm := openai.New()

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Hello, who are you?"),
		),
	)

	err = openaillm.Generate(ctx, t)
	if err != nil {
		panic(err)
	}

	o.Flush(ctx)
}
