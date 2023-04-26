package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	pipe1 := pipeline.NewStep("step1", llm1, decoder.NewDefaultDecoder(), cache)

	myout := &struct {
		First  string
		Second string
	}{}
	prompt2, _ := prompt.NewPromptTemplate(
		`It seems you are a random word generator. Your message '{{.output}}' is nonsense. 
		Anyway I'm fine {{.value}}!`,
		map[string]string{
			"value": "thanks",
		},
	)
	llm2 := pipeline.Llm{
		LlmEngine: &llmmock.JsonLllMock{},
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt2,
	}
	pipe2 := pipeline.NewStep("step2", llm2, decoder.NewJSONDecoder(myout), cache)

	regexDecoder := decoder.NewRegExDecoder(`(\w+)\s(\w+)\s(.*)`)
	prompt3, _ := prompt.NewPromptTemplate(
		`Oh! It seems you are a random JSON word generator. You generated two strings, 
		first:'{{.First}}' and second:'{{.Second}}'. {{.value}}\n\tHowever your first 
		message was: '{{.step1.output}}'`,
		map[string]string{
			"value": "Bye!",
		},
	)
	llm1.Prompt = prompt3
	pipe3 := pipeline.NewStep("step3", llm1, regexDecoder, cache)

	pipelineSteps := pipeline.New(
		pipe1,
		pipe2,
		pipe3,
	)

	response, err := pipelineSteps.Run(context.TODO(), nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Final output: %s\n", strings.Join(response.([]string), ", "))
	fmt.Println("---Memory---")
	dump, _ := json.MarshalIndent(cache.All(), "", "  ")
	fmt.Printf("%s\n", string(dump))
}
