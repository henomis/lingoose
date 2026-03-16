package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
)

func main() {
	openaillm := openai.New().WithModel(openai.GPT4o).WithResponseFormat(openai.ResponseFormatJSONObject).WithMaxTokens(1000)

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Give me a JSON object that describes a person"),
		),
	)

	err := openaillm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
