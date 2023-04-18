package main

import (
	"fmt"

	"github.com/henomis/lingoose/llm"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/prompt/decoder"
)

func main() {

	var output string

	promptTemplate, err := prompt.New(
		map[string]string{
			"text": "How are you?",
		},
		&output,
		decoder.NewStringDecoderFn(),
		newString("{{.text}}"),
	)
	if err != nil {
		panic(err)
	}

	llm := llm.LlmMock{}

	_, err = llm.Completion(promptTemplate)
	if err != nil {
		panic(err)
	}

	fmt.Println(output)

	// ----------

	var matches []string
	promptTemplateRegex, err := prompt.New(
		map[string]string{
			"text": "How are you?",
		},
		&matches,
		decoder.NewRegexDecoderFn(`(\w+), (\w+)`),
		newString("{{.text}}"),
	)
	if err != nil {
		panic(err)
	}

	_, err = llm.Completion(promptTemplateRegex)
	if err != nil {
		panic(err)
	}

	for _, match := range matches {
		fmt.Printf("[%s] ", match)
	}
	fmt.Println()
}

func newString(s string) *string {
	return &s
}
