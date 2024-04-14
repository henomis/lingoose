package main

import (
	"context"
	"fmt"
	"github.com/henomis/lingoose/llm/anthropic"

	"github.com/henomis/lingoose/thread"
)

func main() {
	antropicllm := anthropic.New().WithModel("claude-3-opus-20240229").WithStream(
		func(response string) {
			if response != anthropic.EOS {
				fmt.Print(response)
			} else {
				fmt.Println()
			}
		},
	)

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("How are you?"),
		),
	)

	err := antropicllm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
