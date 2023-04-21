package main

import (
	"encoding/json"
	"fmt"

	"github.com/henomis/lingoose/decoder"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/memory/ram"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	cache := ram.New()

	llmOpenAI, err := openai.New(openai.GPT3TextDavinci003, openai.DefaultOpenAITemperature, openai.DefaultOpenAIMaxTokens, true)
	if err != nil {
		panic(err)
	}

	llm := pipeline.Llm{
		LlmEngine: llmOpenAI,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt.New("Hello how are you?"),
	}
	pipe1 := pipeline.NewStep(
		"step1",
		llm,
		decoder.NewDefaultDecoder(),
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
	pipe2 := pipeline.NewStep(
		"step2",
		llm,
		decoder.NewDefaultDecoder(),
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
	pipe3 := pipeline.NewStep(
		"step3",
		llm,
		decoder.NewDefaultDecoder(),
		cache,
	)

	pipelineSteps := pipeline.New(
		pipe1,
		pipe2,
		pipe3,
	)

	response, err := pipelineSteps.Run(nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("\n\nFinal output: %#v\n\n", response)

	fmt.Println("---Memory---")
	dump, _ := json.MarshalIndent(cache.All(), "", "  ")
	fmt.Printf("%s\n", string(dump))
}
