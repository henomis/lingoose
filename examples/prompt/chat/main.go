package main

import (
	"fmt"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	prompt1, err := prompt.NewPromptTemplate(
		"Translating from {{.input_language}} to {{.output_language}}",
		map[string]string{
			"input_language":  "English",
			"output_language": "French",
		},
	)
	if err != nil {
		panic(err)
	}

	prompt2, err := prompt.NewPromptTemplate(
		"{{.text}}",
		map[string]string{
			"text": "I love programming.",
		},
	)
	if err != nil {
		panic(err)
	}

	chatTemplate := chat.New(
		chat.PromptMessage{
			Type:   chat.MessageTypeSystem,
			Prompt: prompt1,
		},
		chat.PromptMessage{
			Type:   chat.MessageTypeUser,
			Prompt: prompt2,
		},
	)

	messages, err := chatTemplate.ToMessages()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", messages)

}
