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

	m := tool.NewMath()
	s := tool.NewStrLen()
	r := tool.NewDuckDuckGo()

	agent, err := agent.New(openai.NewCompletion().WithStop([]string{"Observation:"}).WithVerbose(true), []agent.Tool{m, s, r})
	if err != nil {
		panic(err)
	}

	res, err := agent.Run(context.Background(), types.M{"question": "i want to do a strange calculation. The population of USA divided by the number of cars in the world."})
	if err != nil {
		panic(err)
	}

	fmt.Println(res["output"])

}
