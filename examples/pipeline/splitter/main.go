package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/decoder"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	llmOpenAI, err := openai.New(openai.GPT3TextDavinci003, openai.DefaultOpenAITemperature, openai.DefaultOpenAIMaxTokens, true)
	if err != nil {
		panic(err)
	}

	llm := pipeline.Llm{
		LlmEngine: llmOpenAI,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt.New("Hello how are you?"),
	}
	step1 := pipeline.NewStep(
		"step1",
		llm,
		decoder.NewDefaultDecoder(),
		nil,
	)

	prompt2, _ := prompt.NewPromptTemplate(
		"Consider the following sentence.\n\nSentence:\n{{.output}}\n\n"+
			"Translate it in {{.language}}!",
		nil,
	)
	llm.Prompt = prompt2
	step2 := pipeline.NewSplitter(
		"step2",
		llm,
		decoder.NewDefaultDecoder(),
		nil,
		func(input interface{}) ([]interface{}, error) {
			return []interface{}{
				mergeMaps(input.(map[string]interface{}), map[string]interface{}{
					"language": "italian",
				}),
				mergeMaps(input.(map[string]interface{}), map[string]interface{}{
					"language": "spanish",
				}),
				mergeMaps(input.(map[string]interface{}), map[string]interface{}{
					"language": "finnish",
				}),
				mergeMaps(input.(map[string]interface{}), map[string]interface{}{
					"language": "french",
				}),
				mergeMaps(input.(map[string]interface{}), map[string]interface{}{
					"language": "german",
				}),
			}, nil
		},
	)

	prompt3, _ := prompt.NewPromptTemplate(
		"For each of the following sentences, detect the language.\n\nSentences:\n{{.output}}\n\n",
		nil,
	)
	llm.Prompt = prompt3
	step3 := pipeline.NewFunnel(
		"step3",
		llm,
		decoder.NewDefaultDecoder(),
		nil,
		func(input []map[string]interface{}) (interface{}, error) {

			outputString := ""

			for _, sentence := range input {
				outputString += "- " + sentence["output"].(string) + "\n"
			}

			return map[string]interface{}{
				"output": outputString,
			}, nil
		},
	)

	pipeLine := pipeline.New(
		step1,
		step2,
		step3,
	)

	response, err := pipeLine.Run(context.Background(), nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Final output: %#v\n", response)

}

func mergeMaps(m1 map[string]interface{}, m2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}
