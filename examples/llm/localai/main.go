package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
	goopenai "github.com/sashabaranov/go-openai"
)

func main() {
	customConfig := goopenai.DefaultConfig("YOUR_API_KEY")
	customConfig.BaseURL = "http://localhost:8080"
	customClient := goopenai.NewClientWithConfig(customConfig)

	openaillm := openai.New().WithClient(customClient).WithModel("ggml-gpt4all-j")

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("What's the NATO purpose?"),
		),
	)

	err := openaillm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
