package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rsest/lingoose/history"
	"github.com/rsest/lingoose/legacy/chat"
	"github.com/rsest/lingoose/legacy/pipeline"
	"github.com/rsest/lingoose/legacy/prompt"
	"github.com/rsest/lingoose/llm/openai"
	"github.com/rsest/lingoose/types"
)

func main() {

	history := history.NewHistoryRAM()

	llmChatOpenAI := openai.NewChat()

	prompt1 := prompt.NewPromptTemplate(
		"You are a {{.mode}} {{.role}}").WithInputs(
		map[string]string{
			"mode": "professional",
		},
	)
	prompt2 := prompt.NewPromptTemplate(
		"Write a {{.length}} joke about a {{.animal}}.").WithInputs(
		map[string]string{
			"length": "short",
		},
	)
	chat := chat.New(
		chat.PromptMessage{
			Type:   chat.MessageTypeSystem,
			Prompt: prompt1,
		},
		chat.PromptMessage{
			Type:   chat.MessageTypeUser,
			Prompt: prompt2,
		},
	)

	llm1 := pipeline.Llm{
		LlmEngine: llmChatOpenAI,
		LlmMode:   pipeline.LlmModeChat,
		Chat:      chat,
	}
	tube1 := pipeline.NewTube(llm1).WithHistory(history)

	pipe := pipeline.New(tube1)

	values := types.M{
		"role":   "joke writer",
		"animal": "goose",
	}
	response, err := pipe.Run(context.Background(), values)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Final output: %s\n", response["output"])
	fmt.Println("---History---")
	dump, _ := json.MarshalIndent(history.All(), "", "  ")
	fmt.Printf("%s\n", string(dump))

}
