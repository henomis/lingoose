package main

import (
	llmmock "github.com/henomis/lingoose/llm/mock"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	prompt := prompt.New("How are you?")

	llm := llmmock.LlmMock{}

	output, err := llm.Completion(prompt.Prompt())
	if err != nil {
		panic(err)
	}

	println(output)
}
