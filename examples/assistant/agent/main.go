package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/assistant"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/observer/langfuse"
	"github.com/henomis/lingoose/thread"

	pythontool "github.com/henomis/lingoose/tool/python"
	serpapitool "github.com/henomis/lingoose/tool/serpapi"
)

func main() {
	ctx := context.Background()

	langfuseObserver := langfuse.New(ctx)
	trace, err := langfuseObserver.Trace(&observer.Trace{Name: "Average Temperature calculator"})
	if err != nil {
		panic(err)
	}

	ctx = observer.ContextWithObserverInstance(ctx, langfuseObserver)
	ctx = observer.ContextWithTraceID(ctx, trace.ID)

	auto := "auto"
	myAssistant := assistant.New(
		openai.New().WithModel(openai.GPT4o).WithToolChoice(&auto).WithTools(
			pythontool.New(),
			serpapitool.New(),
		),
	).WithParameters(
		assistant.Parameters{
			AssistantName:     "AI Assistant",
			AssistantIdentity: "a helpful assistant",
			AssistantScope:    "answering questions",
		},
	).WithThread(
		thread.New().AddMessages(
			thread.NewUserMessage().AddContent(
				thread.NewTextContent("Search the current temperature of New York, Rome, and Tokyo, then calculate the average temperature in Celsius."),
			),
		),
	).WithMaxIterations(10)

	err = myAssistant.Run(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(myAssistant.Thread())

	langfuseObserver.Flush(ctx)
}
