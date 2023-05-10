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
	r := tool.NewDuckDuckGo()

	agent, err := agent.New("test", llmOpenAI, []agent.Tool{m, s, r})
	if err != nil {
		panic(err)
	}

	res, err := agent.Run(context.Background(), types.M{"question": "what was the high temperature in SF yesterday in Fahrenheit? And the same value in celsius?"})
	if err != nil {
		panic(err)
	}

	fmt.Println(res["output"])

}
