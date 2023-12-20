package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
)

func main() {
	openaillm := openai.New(
		openai.GPT3Dot5Turbo0613,
		0,
		256,
		false,
	)

	t := thread.NewThread().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Hello, I'm a user"),
		).AddContent(
			thread.NewTextContent("Can you greet me?"),
		),
	).AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("please reply as a pirate."),
		),
	)

	fmt.Println(t)

	t, err := openaillm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
