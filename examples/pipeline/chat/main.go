package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/henomis/lingoose/legacy/chat"
	"github.com/henomis/lingoose/legacy/decoder"
	"github.com/henomis/lingoose/legacy/memory/ram"
	"github.com/henomis/lingoose/legacy/pipeline"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
)

func main() {

	cache := ram.New()

	llmChatOpenAI := openai.NewChat().WithVerbose(true)
	llmOpenAI := openai.NewCompletion().WithVerbose(true)

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
	tube1 := pipeline.NewTube(llm1).WithMemory("step1", cache)

	prompt3 := prompt.NewPromptTemplate(
		"Considering the following joke.\n\njoke:\n{{.output}}\n\n{{.command}}").WithInputs(
		map[string]string{
			"command": "Put the joke in a JSON object with only one field called 'joke'. " +
				"Do not add other json fields. Do not add other information.",
		},
	)
	llm2 := pipeline.Llm{
		LlmEngine: llmOpenAI,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt3,
	}

	tube2 := pipeline.NewTube(llm2).WithDecoder(decoder.NewJSONDecoder()).WithMemory("step2", cache)

	pipe := pipeline.New(tube1, tube2)

	values := types.M{
		"role":   "joke writer",
		"animal": "goose",
	}
	response, err := pipe.Run(context.Background(), values)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Final output: %#v\n", response)
	fmt.Println("---Memory---")
	dump, _ := json.MarshalIndent(cache.All(), "", "  ")
	fmt.Printf("%s\n", string(dump))

}
