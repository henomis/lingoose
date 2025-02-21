package main

import (
	"context"

	"github.com/rsest/lingoose/legacy/prompt"
	llmmock "github.com/rsest/lingoose/llm/mock"
)

func main() {

	prompt := prompt.New("How are you?")

	llm := llmmock.LlmMock{}

	output, err := llm.Completion(context.Background(), prompt.String())
	if err != nil {
		panic(err)
	}

	println(output)
}
