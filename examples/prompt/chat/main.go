package main

import (
	"fmt"

	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/prompt/chat"
)

func main() {

	chatTemplate := chat.New(
		[]chat.PromptMessage{
			{
				Type: chat.MessageTypeSystem,
				Prompt: &prompt.Prompt{
					Input: map[string]string{
						"input_language":  "English",
						"output_language": "French",
					},
					OutputDecoder: nil,
					Template:      newString("Translating from {{.input_language}} to {{.output_language}}"),
				},
			},
			{
				Type: chat.MessageTypeUser,
				Prompt: &prompt.Prompt{
					Input: map[string]string{
						"text": "I love programming.",
					},
					OutputDecoder: nil,
					Template:      newString("{{.text}}"),
				},
			},
		},
	)

	messages, err := chatTemplate.ToMessages()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", messages)

}

func newString(s string) *string {
	return &s
}
