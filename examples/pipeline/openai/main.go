package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/memory/ram"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
)

func main() {

	cache := ram.New()

	llmOpenAI, err := openai.NewCompletion()
	if err != nil {
		panic(err)
	}

	llmOpenAI.SetCallback(func(response types.Meta) {
		fmt.Printf("USAGE: %#v\n", response)
	})

	llm := pipeline.Llm{
		LlmEngine: llmOpenAI,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt.New("Hello how are you?"),
	}
	tube1 := pipeline.NewTube(
		"step1",
		llm,
		nil,
		cache,
	)

	prompt2, _ := prompt.NewPromptTemplate(
		"Consider the following sentence.\n\nSentence:\n{{.output}}\n\n"+
			"Translate it in {{.language}}!",
		map[string]string{
			"language": "italian",
		},
	)
	llm.Prompt = prompt2
	tube2 := pipeline.NewTube(
		"step2",
		llm,
		nil,
		nil,
	)

	prompt3, _ := prompt.NewPromptTemplate(
		"Consider the following sentence.\n\nSentence:\n{{.step1.output}}"+
			"\n\nTranslate it in {{.language}}!",
		map[string]string{
			"language": "spanish",
		},
	)
	llm.Prompt = prompt3
	step3 := pipeline.NewTube(
		"step3",
		llm,
		nil,
		cache,
	)

	pipeLine := pipeline.New(
		tube1,
		tube2,
		step3,
	)

	response, err := pipeLine.Run(context.Background(), nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("\n\nFinal output: %#v\n\n", response)

	fmt.Println("---Memory---")
	dump, _ := json.MarshalIndent(cache.All(), "", "  ")
	fmt.Printf("%s\n", string(dump))
}
