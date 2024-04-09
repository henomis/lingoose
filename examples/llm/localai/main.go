package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/localai"
	"github.com/henomis/lingoose/thread"
)

func main() {
	localaillm := localai.New("http://localhost:8080").WithModel("ggml-gpt4all-j")

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("What's the NATO purpose?"),
		),
	)

	err := localaillm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
