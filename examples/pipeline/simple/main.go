package main

import (
	"fmt"
	"strings"

	"github.com/henomis/lingoose/decoder"
	"github.com/henomis/lingoose/llm"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	llm1 := &llm.LlmMock{}
	prompt1 := prompt.New("Hello how are you?")
	pipe1 := pipeline.NewStep("step1", llm1, prompt1, nil, decoder.NewDefaultDecoder(), nil)

	myout := &struct {
		First  string
		Second string
	}{}
	llm2 := &llm.JsonLllMock{}
	prompt2, _ := prompt.NewPromptTemplate(
		"It seems you are a random word generator. Your message '{{.output}}' is nonsense. Anyway I'm fine {{.value}}!",
		map[string]string{
			"value": "thanks",
		},
	)
	pipe2 := pipeline.NewStep("step2", llm2, prompt2, myout, decoder.NewJSONDecoder(), nil)

	var values []string
	prompt3, _ := prompt.NewPromptTemplate(
		"Oh! It seems you are a random JSON word generator. You generated two strings, first:'{{.First}}' and second:'{{.Second}}'. {{.value}}",
		map[string]string{
			"value": "Bye!",
		},
	)
	pipe3 := pipeline.NewStep("step3", llm1, prompt3, values, decoder.NewRegExDecoder(`(\w+?)\s(\w+?)\s(.*)`), nil)

	ovearallPipe := pipeline.New(
		pipe1,
		pipe2,
		pipe3,
	)

	response, err := ovearallPipe.Run(nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Final output: %s\n", strings.Join(response.([]string), ", "))
}
