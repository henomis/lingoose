package main

import (
	"fmt"

	"github.com/henomis/lingoose/decoder"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/memory/ram"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	llmOpenAI, err := openai.New(openai.GPT3TextDavinci003, true)
	if err != nil {
		panic(err)
	}
	cache := ram.New()

	prompt1 := prompt.New("Hello how are you?")
	pipe1 := pipeline.NewStep("step1", llmOpenAI, prompt1, nil, decoder.NewDefaultDecoder(), cache)

	prompt2, _ := prompt.NewPromptTemplate(
		"Consider the following sentence.\n\nSentence:\n{{.output}}\n\nTranslate it in {{.language}}!",
		map[string]string{
			"language": "italian",
		},
	)
	pipe2 := pipeline.NewStep("step2", llmOpenAI, prompt2, nil, decoder.NewDefaultDecoder(), nil)

	prompt3, _ := prompt.NewPromptTemplate(
		"Consider the following sentence.\n\nSentence:\n{{.step1.output}}\n\nTranslate it in {{.language}}!",
		map[string]string{
			"language": "spanish",
		},
	)
	pipe3 := pipeline.NewStep("step3", llmOpenAI, prompt3, nil, decoder.NewDefaultDecoder(), cache)

	pipelineSteps := pipeline.New(
		pipe1,
		pipe2,
		pipe3,
	)

	response, err := pipelineSteps.Run(nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("\nFinal output: %#v\n", response)
}
