package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rsest/lingoose/legacy/memory/ram"
	"github.com/rsest/lingoose/legacy/pipeline"
	"github.com/rsest/lingoose/legacy/prompt"
	"github.com/rsest/lingoose/llm/openai"
	"github.com/rsest/lingoose/types"
)

func main() {

	cache := ram.New()

	llmOpenAI := openai.NewCompletion().WithVerbose(true)

	llmOpenAI.WithCallback(func(response types.Meta) {
		fmt.Printf("USAGE: %#v\n", response)
	}).WithVerbose(true)

	llm := pipeline.Llm{
		LlmEngine: llmOpenAI,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt.New("Hello how are you?"),
	}
	tube1 := pipeline.NewTube(llm).WithMemory("step1", cache)

	prompt2 := prompt.NewPromptTemplate(
		"Consider the following sentence.\n\nSentence:\n{{.output}}\n\n" +
			"Translate it in {{.language}}!").WithInputs(
		map[string]string{
			"language": "italian",
		},
	)
	llm.Prompt = prompt2
	tube2 := pipeline.NewTube(llm)

	prompt3 := prompt.NewPromptTemplate(
		"Consider the following sentence.\n\nSentence:\n{{.step1.output}}" +
			"\n\nTranslate it in {{.language}}!").WithInputs(
		map[string]string{
			"language": "spanish",
		},
	)
	llm.Prompt = prompt3
	step3 := pipeline.NewTube(llm).WithMemory("step3", cache)

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
