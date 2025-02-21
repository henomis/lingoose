package main

import (
	"context"

	"github.com/rsest/lingoose/llm/cohere"
)

func main() {

	llm := cohere.NewCompletion().WithMaxTokens(100).WithTemperature(0.1).WithVerbose(true)

	_, err := llm.Completion(context.Background(), "Where is Rome?")
	if err != nil {
		panic(err)
	}

}
