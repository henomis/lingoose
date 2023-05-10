package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/henomis/lingoose/decoder"
	llmmock "github.com/henomis/lingoose/llm/mock"
	"github.com/henomis/lingoose/memory/ram"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	cache := ram.New()

	prompt1 := prompt.New("Hello how are you?")
	llm1 := pipeline.Llm{
		LlmEngine: &llmmock.LlmMock{},
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt1,
	}
	tube1 := pipeline.NewTube(llm1).WithMemory("step1", cache)

	prompt2, _ := prompt.NewPromptTemplate(
		"It seems you are a random word generator. Your message '{{.output}}' is nonsense. "+
			"Anyway I'm fine {{.value}}!",
		map[string]string{
			"value": "thanks",
		},
	)
	llm2 := pipeline.Llm{
		LlmEngine: &llmmock.JsonLllMock{},
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt2,
	}
	tube2 := pipeline.NewTube(llm2).WithDecoder(decoder.NewJSONDecoder()).WithMemory("step2", cache)

	prompt3, _ := prompt.NewPromptTemplate(
		"Oh! It seems you are a random JSON word generator. You generated two strings, "+
			"first:'{{.step2.output.first}}' and second:'{{.step2.output.second}}'. {{.value}}\n\nHowever your first "+
			"message was: '{{.step1.output}}'",
		map[string]string{
			"value": "Bye!",
		},
	)
	llm1.Prompt = prompt3
	tube3 := pipeline.NewTube(llm1).WithDecoder(decoder.NewRegExDecoder(`(\w+)\s(\w+)\s(.*)`)).WithMemory("step3", cache)

	prompt4, _ := prompt.NewPromptTemplate("Well here is your answer: "+
		"{{ range  $value := .step3.output }}[{{$value}}] {{end}}", nil)
	llm1.Prompt = prompt4
	tube4 := pipeline.NewTube(llm1).WithMemory("step4", cache)

	pipelineTubes := pipeline.New(
		tube1,
		tube2,
		tube3,
		tube4,
	)

	response, err := pipelineTubes.Run(context.Background(), nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Final output: %#v\n", response)
	fmt.Println("---Memory---")
	dump, _ := json.MarshalIndent(cache.All(), "", "  ")
	fmt.Printf("%s\n", string(dump))
}
