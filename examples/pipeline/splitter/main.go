package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
)

func main() {

	llmOpenAI := openai.NewCompletion().WithVerbose(true)

	llm := pipeline.Llm{
		LlmEngine: llmOpenAI,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt.New("Hello how are you?"),
	}
	tube1 := pipeline.NewTube(llm)

	prompt2 := prompt.NewPromptTemplate(
		"Consider the following sentence.\n\nSentence:\n{{.output}}\n\n" +
			"Translate it in {{.language}}!")

	llm.Prompt = prompt2
	tube2 := pipeline.NewSplitter(
		"step2",
		llm,
		nil,
		nil,
		func(input types.M) ([]types.M, error) {
			return []types.M{
				mergeMaps(input, types.M{
					"language": "italian",
				}),
				mergeMaps(input, types.M{
					"language": "spanish",
				}),
				mergeMaps(input, types.M{
					"language": "finnish",
				}),
				mergeMaps(input, types.M{
					"language": "french",
				}),
				mergeMaps(input, types.M{
					"language": "german",
				}),
			}, nil
		},
	)

	prompt3 := prompt.NewPromptTemplate(
		"For each of the following sentences, detect the language.\n\nSentences:\n" +
			"{{ range $i, $key := .output }}{{ $i }}. {{ $key.output }}\n{{ end }}\n\n",
	)
	llm.Prompt = prompt3
	tube3 := pipeline.NewTube(llm)

	pipeLine := pipeline.New(
		tube1,
		tube2,
		tube3,
	)

	response, err := pipeLine.Run(context.Background(), nil)
	if err != nil {
		fmt.Println(err)
	}

	data, _ := json.MarshalIndent(response, "", "  ")

	fmt.Printf("Final output: %s\n", data)

}

func mergeMaps(m1 types.M, m2 types.M) types.M {
	merged := make(types.M)
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}
