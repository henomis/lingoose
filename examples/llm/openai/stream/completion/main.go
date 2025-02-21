package main

import (
	"context"
	"fmt"

	"github.com/rsest/lingoose/llm/openai"
)

func main() {

	llm := openai.NewCompletion()

	err := llm.CompletionStream(context.Background(), func(output string) {
		fmt.Printf("%s", output)
	}, "Tell me a joke about geese")
	if err != nil {
		panic(err)
	}

	fmt.Println()

}
