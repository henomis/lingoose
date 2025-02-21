package main

import (
	"context"

	"github.com/rsest/lingoose/llm/huggingface"
)

func main() {

	llm := huggingface.New("microsoft/DialoGPT-medium", 1.0, true)

	_, err := llm.Completion(context.Background(), "What is the NATO purpose?")
	if err != nil {
		panic(err)
	}

}
