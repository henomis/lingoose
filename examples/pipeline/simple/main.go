package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/decoder"
	llmmock "github.com/henomis/lingoose/llm/mock"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	llm1 := pipeline.Llm{
		LlmEngine: &llmmock.LlmMock{},
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt.New("Hello how are you?"),
	}
	tube1 := pipeline.NewTube("step1", llm1, nil, nil)

	prompt2, _ := prompt.NewPromptTemplate(
		"It seems you are a random word generator. Your message '{{.output}}' is nonsense. Anyway I'm fine {{.value}}!",
		map[string]string{
			"value": "thanks",
		},
	)
	llm2 := pipeline.Llm{
		LlmEngine: &llmmock.JsonLllMock{},
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt2,
	}
	tube2 := pipeline.NewTube("step2", llm2, decoder.NewJSONDecoder(), nil)

	prompt3, _ := prompt.NewPromptTemplate(
		"Oh! It seems you are a random JSON word generator. You generated two strings, first:'{{.First}}' and second:'{{.Second}}'. {{.value}}",
		map[string]string{
			"value": "Bye!",
		},
	)
	llm1.Prompt = prompt3
	tube3 := pipeline.NewTube("step3", llm1, decoder.NewRegExDecoder(`(\w+?)\s(\w+?)\s(.*)`), nil)

	pipelineTubes := pipeline.New(
		tube1,
		tube2,
		tube3,
	)

	response, err := pipelineTubes.Run(context.Background(), nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Final output: %#v\n", response)
}
