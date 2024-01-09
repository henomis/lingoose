package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/ollama"
	"github.com/henomis/lingoose/thread"
)

func main() {
	t := thread.New()
	t.AddMessage(thread.NewUserMessage().AddContent(
		thread.NewTextContent("Hello"),
	))

	err := ollama.New().WithEndpoint("http://localhost:11434/api").WithModel("llama2").
		WithStream(true, func(s string) {
			fmt.Print(s)
		}).Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t.String())

}
