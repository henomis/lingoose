package main

import (
	"context"

	"github.com/henomis/lingoose/llm/huggingface"
)

func main() {

	llm := huggingface.New("gpt2", 0.1, true).WithMode(huggingface.ModeTextGeneration)

	_, err := llm.Completion(context.Background(), "What is the NATO purpose?")
	if err != nil {
		panic(err)
	}

	_, err = llm.BatchCompletion(
		context.Background(),
		[]string{
			"Write a joke about geese.",
			"What is the NATO purpose?",
		},
	)
	if err != nil {
		panic(err)
	}

}
