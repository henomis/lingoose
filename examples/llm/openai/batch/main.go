package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/openai"
)

func main() {

	llm := openai.NewCompletion()

	outputs, err := llm.BatchCompletion(
		context.Background(),
		[]string{
			"Tell me a joke about geese",
			"What is the NATO purpose?",
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", outputs)

	var output1, output2 string

	err = llm.BatchCompletionStream(
		context.Background(),
		[]openai.StreamCallback{
			func(output string) {
				fmt.Printf("{%s}", output)
				output1 += output
			},
			func(output string) {
				fmt.Printf("[%s]", output)
				output2 += output
			},
		},
		[]string{
			"Tell me a joke about geese",
			"What is the NATO purpose?",
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("------")
	fmt.Printf("%s\n", output1)
	fmt.Printf("%s\n", output2)

}
