package main

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/llm/anthropic"
	"github.com/henomis/lingoose/thread"
)

func main() {
	anthropicllm := anthropic.NewAnthropic(os.Getenv("ANTHROPIC_API_KEY")).WithModel(anthropic.ModelClaude_3_Opus_20240229).WithStream(true,
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

	err := anthropicllm.Chat(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
