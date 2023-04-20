package main

import (
	"fmt"

	"github.com/henomis/lingoose/llm"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/prompt/decoder"
)

func main() {

	llm1 := &llm.LlmMock{}
	prompt1 := prompt.New("ciao come stai?")
	pipe1 := pipeline.NewStep(llm1, prompt1, nil, decoder.NewDefaultDecoder())

	myout := &struct {
		First  string
		Second string
	}{}
	llm2 := &llm.JsonLllMock{}
	prompt2, _ := prompt.NewPromptTemplate(
		"basato su '{{.output}}', sto bene {{.saluti}}",
		map[string]string{
			"saluti": "ciao",
		},
	)
	pipe2 := pipeline.NewStep(llm2, prompt2, myout, decoder.NewJSONDecoder())

	var values []string
	prompt3, _ := prompt.NewPromptTemplate(
		"basato su '{{.First}}' e soprattutto su '{{.Second}}', sto bene {{.saluti}}",
		map[string]string{
			"saluti": "ciao",
		},
	)
	pipe3 := pipeline.NewStep(llm1, prompt3, values, decoder.NewRegExDecoder(`(\w+)\s(\w+)\s(.*)`))

	ovearallPipe := pipeline.New(
		pipe1,
		pipe2,
		pipe3,
	)

	response, err := ovearallPipe.Run(nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%#v\n", response)
}
