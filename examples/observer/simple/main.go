package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/thread"
)

func main() {
	observer := observer.NewSimpleObserver()
	openaillm := openai.New().WithObserver(observer)

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Hello, world!"),
		),
	)

	err := openaillm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
	observer.Wait(context.Background())
}
