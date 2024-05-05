package main

import (
	"context"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/observer/langfuse"
	"github.com/henomis/lingoose/thread"
)

func main() {
	ctx := context.Background()

	o := langfuse.New(ctx)
	trace, err := o.Trace(&observer.Trace{Name: "Who are you"})
	if err != nil {
		panic(err)
	}

	span, err := o.Span(
		&observer.Span{
			TraceID: trace.ID,
			Name:    "SPAN",
		},
	)
	if err != nil {
		panic(err)
	}

	ctx = observer.StoreParentIDInContext(ctx, span.ID)

	openaillm := openai.New().WithObserver(o, trace.ID)

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
