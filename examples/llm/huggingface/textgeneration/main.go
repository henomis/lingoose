package main

import (
	"context"

	"github.com/henomis/lingoose/llm/huggingface"
)

func main() {

	llm := huggingface.New("gpt2", 0.7, true).WithMode(huggingface.HuggingFaceModeTextGeneration)

	_, err := llm.Completion(context.Background(), "What is the NATO purpose?")
	if err != nil {
		panic(err)
	}

}
