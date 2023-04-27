package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/decoder"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/memory/ram"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
)

func main() {

	cache := ram.New()

	llmChatOpenAI, err := openai.New(openai.GPT3Dot5Turbo, openai.DefaultOpenAITemperature, openai.DefaultOpenAIMaxTokens, true)
	if err != nil {
		panic(err)
	}

	llmOpenAI, err := openai.New(openai.GPT3TextDavinci002, openai.DefaultOpenAITemperature, openai.DefaultOpenAIMaxTokens, true)
	if err != nil {
		panic(err)
	}

	prompt1, _ := prompt.NewPromptTemplate(
		"You are a {{.mode}} {{.role}}",
		map[string]string{
			"mode": "professional",
		},
	)
	prompt2, _ := prompt.NewPromptTemplate(
		"Write a {{.length}} joke about a {{.animal}}.",
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
	tube1 := pipeline.NewTube(
		"step1",
		llm1,
		nil,
		cache,
	)

	prompt3, _ := prompt.NewPromptTemplate(
		"Considering the following joke.\n\njoke:\n{{.output}}\n\n{{.command}}",
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

	tube2 := pipeline.NewTube(
		"step2",
		llm2,
		decoder.NewJSONDecoder(),
		cache,
	)

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
