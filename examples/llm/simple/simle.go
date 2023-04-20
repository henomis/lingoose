package main

import (
	"github.com/henomis/lingoose/llm"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	prompt := prompt.New("How are you?")

	llm := llm.LlmMock{}

	output, err := llm.Completion(prompt)
	if err != nil {
		panic(err)
	}

	println(output)
}
