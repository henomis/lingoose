package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/legacy/chat"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	chat := chat.New(
		chat.PromptMessage{
			Type:   chat.MessageTypeSystem,
			Prompt: prompt.New("You are a professional joke writer"),
		},
		chat.PromptMessage{
			Type:   chat.MessageTypeUser,
			Prompt: prompt.New("Write a joke about a goose"),
		},
	)

	llmOpenAI := openai.NewLegacy(openai.GPT3Dot5Turbo, openai.DefaultOpenAITemperature, openai.DefaultOpenAIMaxTokens, true)

	response, err := llmOpenAI.Chat(context.Background(), chat)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n%#v", response)

}
