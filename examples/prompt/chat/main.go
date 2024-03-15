package main

import (
	"fmt"

	"github.com/henomis/lingoose/legacy/chat"
	"github.com/henomis/lingoose/legacy/prompt"
)

func main() {

	prompt1 := prompt.NewPromptTemplate(
		"Translating from {{.input_language}} to {{.output_language}}").WithInputs(
		map[string]string{
			"input_language":  "English",
			"output_language": "French",
		},
	)

	prompt2 := prompt.NewPromptTemplate(
		"{{.text}}").WithInputs(
		map[string]string{
			"text": "I love programming.",
		},
	)

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
