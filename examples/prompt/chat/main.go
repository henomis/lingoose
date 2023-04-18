package main

import (
	"fmt"

	"github.com/henomis/lingoose/prompt/chat"
	"github.com/henomis/lingoose/prompt/template"
)

func main() {

	chatTemplate := chat.New(
		[]chat.MessageTemplate{
			{
				Type: chat.MessageTypeSystem,
				Template: template.New(
					[]string{"input_language", "output_language"},
					[]string{},
					"You are a helpful assistant that translates {{.input_language}} to {{.output_language}}.",
					nil,
				),
			},
			{
				Type: chat.MessageTypeUser,
				Template: template.New(
					[]string{"text"},
					[]string{},
					"{{.text}}",
					nil,
				),
			},
		},
	)

	messages, err := chatTemplate.ToMessages(
		template.Inputs{
			"input_language":  "English",
			"output_language": "French",
			"text":            "I love programming.",
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", messages)

}
