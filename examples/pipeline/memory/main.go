package main

import (
	"encoding/json"
	"fmt"

	"github.com/henomis/lingoose/decoder"
	"github.com/henomis/lingoose/llm"
	"github.com/henomis/lingoose/memory"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	cache := memory.NewSimpleMemory()

	llm1 := &llm.LlmMock{}
	prompt1 := prompt.New("ciao come stai?")
	pipe1 := pipeline.NewStep("step1", llm1, prompt1, nil, decoder.NewDefaultDecoder(), cache)

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
	pipe2 := pipeline.NewStep("step2", llm2, prompt2, myout, decoder.NewJSONDecoder(), cache)

	var values []string
	prompt3, _ := prompt.NewPromptTemplate(
		"basato su '{{.First}}' e soprattutto su '{{.Second}}', sto bene {{.saluti}}. Primo passo: {{.step1.output}}",
		map[string]string{
			"saluti": "ciao",
		},
	)
	pipe3 := pipeline.NewStep("step3", llm1, prompt3, values, decoder.NewRegExDecoder(`(\w+)\s(\w+)\s(.*)`), cache)

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
	fmt.Println("---Memory---")
	dump, _ := json.MarshalIndent(cache.All(), "", "  ")
	fmt.Printf("%s\n", string(dump))
}
