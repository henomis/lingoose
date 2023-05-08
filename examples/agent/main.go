package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/agent"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/tool"
	"github.com/henomis/lingoose/types"
)

func main() {

	llmOpenAI, err := openai.New(openai.GPT3TextDavinci003, openai.DefaultOpenAITemperature, openai.DefaultOpenAIMaxTokens, true)
	if err != nil {
		panic(err)
	}

	m := tool.NewMath()
	s := tool.NewStrLen()

	agent, err := agent.New("test", llmOpenAI, []agent.Tool{m, s})
	if err != nil {
		panic(err)
	}

	res, err := agent.Run(context.Background(), types.M{"question": "What is the sghent of the string \"hello world\" multiplied by 5?"})
	if err != nil {
		panic(err)
	}

	fmt.Println(res["output"])

}
