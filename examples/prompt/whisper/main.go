package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/prompt"
)

type Inputs struct {
	Name string `json:"name"`
}

func main() {

	var input Inputs
	input.Name = "world"

	promptTemplate, err := prompt.NewPromptFromAudioFile(
		context.Background(),
		"hello.mp3",
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(promptTemplate.Prompt())

}
