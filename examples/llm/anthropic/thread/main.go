package main

import (
	"context"
	"fmt"
	"github.com/henomis/lingoose/llm/anthropic"

	"github.com/henomis/lingoose/thread"
)

type Answer struct {
	Answer string `json:"answer" jsonschema:"description=the pirate answer"`
}

func getAnswer(a Answer) string {
	return "🦜 ☠️ " + a.Answer
}

func newStr(str string) *string {
	return &str
}

func main() {
	anthropicllm := anthropic.New().WithModel(anthropic.Claude3Sonnet)
	err := anthropicllm.BindFunction(
		getAnswer,
		"getPirateAnswer",
		"use this function to get the pirate answer",
	)
	if err != nil {
		panic(err)
	}
	//anthropicllm.WithStream(func(a string) {
	//	if a == openai.EOS {
	//		fmt.Printf("\n")
	//		return
	//	}
	//	fmt.Printf("%s", a)
	//})

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Hello, I'm a user"),
		).AddContent(
			thread.NewTextContent("Can you greet me?"),
		).AddContent(
			thread.NewTextContent("please greet me as a pirate."),
		),
	)

	fmt.Println(t)

	err = anthropicllm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	t.AddMessage(thread.NewUserMessage().AddContent(
		thread.NewTextContent("now translate to italian as a poem"),
	))

	fmt.Println(t)
	// disable functions
	err = anthropicllm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
