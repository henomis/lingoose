package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func main() {
	type Result struct {
		FirstName string
		LastName  string
		Age       int
		Job       string
	}
	var result Result
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		panic(err)
	}

	openaillm := openai.New().
		WithModel("gpt-4.1-nano").
		WithResponseFormat(openai.ResponseFormatTypeJSONSchema).
		WithResponseFormatJSONSchema(&openai.ResponseFormatJSONSchema{
			Name:   "Person",
			Schema: schema,
			Strict: true,
		}).
		WithMaxTokens(1000)

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("Give me a JSON object that describes a person"),
		),
	)

	err = openaillm.Generate(context.Background(), t)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}
