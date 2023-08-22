package main

import (
	"context"
	"fmt"

	goopenai "github.com/sashabaranov/go-openai"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	chat := chat.New(
		chat.PromptMessage{
			Type:   chat.MessageTypeUser,
			Prompt: prompt.New("How are you?"),
		},
	)

	customConfig := goopenai.DefaultConfig("")
	customConfig.BaseURL = "http://localhost:8080"
	customClient := goopenai.NewClientWithConfig(customConfig)

	llm := openai.NewChat().WithClient(customClient).WithModel("ggml-gpt4all-j")

	err := llm.ChatStream(context.Background(), func(output string) {
		fmt.Printf("%s", output)
	}, chat)
	if err != nil {
		panic(err)
	}

	fmt.Println()

}
