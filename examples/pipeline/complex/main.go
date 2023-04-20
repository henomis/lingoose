package main

import (
	"fmt"

	"github.com/henomis/lingoose/llm"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/prompt/decoder"
)

func main() {

	var output []string `json:"output"`
	promptTemplate, err := prompt.New(
		"How are you?",
		&output,
		decoder.NewRegexDecoderFn(`(.*)`),
		nil,
	)
	if err != nil {
		panic(err)
	}
	llm := &llm.LlmMock{}
	pipeline1 := pipeline.New(llm, promptTemplate)

	var output2 []string
	promptTemplate2, err := prompt.New(
		nil,
		&output2,
		decoder.NewRegexDecoderFn(`(\w+), (\w+)`),
		newString("Given the previous contex {{.output}}, what is your name?"),
	)
	if err != nil {
		panic(err)
	}
	pipeline2 := pipeline.New(llm, promptTemplate2)

	var pipelines pipeline.Pipelines
	pipelines = append(pipelines, *pipeline1)
	pipelines = append(pipelines, *pipeline2)

	result, err := pipelines.Run()
	if err != nil {
		panic(err)
	}
	_ = result

	fmt.Printf("%#v", output2)

}

func newString(s string) *string {
	return &s
}
