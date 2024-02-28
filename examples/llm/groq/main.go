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
	customConfig.BaseURL = "https://api.groq.com/openai/v1"
	customClient := goopenai.NewClientWithConfig(customConfig)

	openaillm := openai.New().WithClient(customClient).WithModel("mixtral-8x7b-32768")

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
