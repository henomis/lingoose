package main

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/legacy/decoder"
	"github.com/henomis/lingoose/legacy/pipeline"
	llmmock "github.com/henomis/lingoose/llm/mock"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	llm1 := pipeline.Llm{
		LlmEngine: &llmmock.LlmMock{},
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt.New("Hello how are you?"),
	}
	tube1 := pipeline.NewTube(llm1)

	prompt2 := prompt.NewPromptTemplate(
		"It seems you are a random word generator. Your message '{{.output}}' is nonsense. Anyway I'm fine {{.value}}!").WithInputs(
		map[string]string{
			"value": "thanks",
		},
	)
	llm2 := pipeline.Llm{
		LlmEngine: &llmmock.JSONLllMock{},
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt2,
	}
	tube2 := pipeline.NewTube(llm2).WithDecoder(decoder.NewJSONDecoder())

	prompt3 := prompt.NewPromptTemplate(
		"Oh! It seems you are a random JSON word generator. You generated two strings, first:'{{.output.first}}' and second:'{{.output.second}}'. {{.value}}").WithInputs(
		map[string]string{
			"value": "Bye!",
		},
	)
	llm1.Prompt = prompt3
	tube3 := pipeline.NewTube(llm1).WithDecoder(decoder.NewRegExDecoder(`(\w+?)\s(\w+?)\s(.*)`))

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
