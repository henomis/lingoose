package main

import (
	"context"
	"fmt"

	"github.com/rsest/lingoose/llm/cohere"
	"github.com/rsest/lingoose/thread"
)

func main() {
	t := thread.New()
	t.AddMessage(thread.NewUserMessage().AddContent(
		thread.NewTextContent("Hello"),
	))

	err := cohere.New().WithMaxTokens(1000).WithTemperature(0).
		WithStream(
			func(s string) {
				fmt.Print(s)
			},
		).Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t.String())

}
