package main

import (
	"context"
	"fmt"

	"github.com/rsest/lingoose/llm/groq"
	"github.com/rsest/lingoose/thread"
)

func main() {
	// The Groq API key is expected to be set in the GROQ_API_KEY environment variable
	groqllm := groq.New().WithModel("mixtral-8x7b-32768")

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("What's the NATO purpose?"),
		),
	)

	err := groqllm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
