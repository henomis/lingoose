package main

import (
	"fmt"

	"github.com/henomis/lingoose/chat"
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
			Prompt: prompt.New("Write a joke about a cat"),
		},
	)

	llmOpenAI, err := openai.New(openai.GPT3Dot5Turbo, true)
	if err != nil {
		panic(err)
	}

	response, err := llmOpenAI.Chat(chat)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n%#v", response)

}
