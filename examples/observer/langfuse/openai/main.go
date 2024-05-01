package main

import (
	"context"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/observer/langfuse"
	"github.com/henomis/lingoose/thread"
)

func main() {
	o := langfuse.New(context.Background())
	trace, err := o.Trace(&observer.Trace{Name: "Who are you"})
	if err != nil {
		panic(err)
	}

	openaillm := openai.New().WithObserver(o, trace.ID)

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Hello, who are you?"),
		),
	)

	err = openaillm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	o.Flush(context.Background())
}
