package main

import (
	"fmt"

	"github.com/henomis/lingoose/llm"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/prompt/decoder"
)

func main() {

	llm := &llm.LlmMock{}
	simpleDecoder := decoder.NewSimpleDecoder()

	prompt1 := prompt.New("ciao come stai?")
	pipe1 := pipeline.New(llm, prompt1, simpleDecoder)

	prompt2, _ := prompt.NewPromptTemplate(
		"basato su '{{.output}}', sto bene {{.saluti}}",
		map[string]string{
			"saluti": "ciao",
		},
	)
	pipe2 := pipeline.New(llm, prompt2, simpleDecoder)

	ovearallPipe := pipeline.Pipelines{
		*pipe1,
		*pipe2,
	}

	response, err := ovearallPipe.Run(nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(response)

}

func newString(s string) *string {
	return &s
}
