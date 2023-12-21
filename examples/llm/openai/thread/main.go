package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
)

type Answer struct {
	Answer string `json:"answer" jsonschema:"description=the pirate answer"`
}

func getAnswer(a Answer) string {
	return a.Answer
}

func newStr(str string) *string {
	return &str
}

func main() {
	openaillm := openai.New()
	openaillm.WithToolChoice(newStr("getPirateAnswer"))
	openaillm.BindFunction(
		getAnswer,
		"getPirateAnswer",
		"use this function to get the pirate answer",
	)

	t := thread.NewThread().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Hello, I'm a user"),
		).AddContent(
			thread.NewTextContent("Can you greet me?"),
		),
	).AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("please greet me as a pirate."),
		),
	)

	fmt.Println(t)

	err := openaillm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	t.AddMessage(thread.NewUserMessage().AddContent(
		thread.NewTextContent("now translate to italian as a poem"),
	))

	fmt.Println(t)
	// disable functions
	openaillm.WithToolChoice(nil)

	err = openaillm.Stream(context.Background(), t, func(a string) {
		if a == openai.EOS {
			fmt.Printf("\n")
			return
		}
		fmt.Printf("%s", a)
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
