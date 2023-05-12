package main

import (
	"context"

	"github.com/henomis/lingoose/llm/llamacpp"
)

func main() {

	llm := llamacpp.NewCompletion().WithMaxTokens(10).WithTemperature(0.1).WithVerbose(true)

	_, err := llm.Completion(context.Background(), "Where is Rome?")
	if err != nil {
		panic(err)
	}

}
