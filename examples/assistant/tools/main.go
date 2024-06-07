package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/assistant"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/observer/langfuse"
	"github.com/henomis/lingoose/thread"

	humantool "github.com/henomis/lingoose/tool/human"
	pythontool "github.com/henomis/lingoose/tool/python"
	serpapitool "github.com/henomis/lingoose/tool/serpapi"
)

func main() {
	ctx := context.Background()

	langfuseObserver := langfuse.New(ctx)
	trace, err := langfuseObserver.Trace(&observer.Trace{Name: "Italian guests calculator"})
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
			humantool.New(),
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
				thread.NewTextContent("search the top 3 italian dishes and then their costs, then ask the user's budget in euros and calculate how many guests can be invited for each dish"),
			),
		),
	).WithToolsMaxIterations(10)

	err = myAssistant.Run(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(myAssistant.Thread())

	langfuseObserver.Flush(ctx)
}
