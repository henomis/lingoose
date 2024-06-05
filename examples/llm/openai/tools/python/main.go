package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/tool/python"
)

func main() {
	newStr := func(str string) *string {
		return &str
	}
	llm := openai.New().WithModel(openai.GPT3Dot5Turbo0613).WithToolChoice(newStr("auto")).WithTools(
		python.New(),
	)

	t := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent("calculate reverse string of 'ailatiditalia', don't try to guess, let's use appropriate tool"),
		),
	)

	llm.Generate(context.Background(), t)
	if t.LastMessage().Role == thread.RoleTool {
		llm.Generate(context.Background(), t)
	}

	fmt.Println(t)
}
