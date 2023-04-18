package main

import (
	"fmt"

	"github.com/henomis/lingopipes/prompt/template"
)

func main() {

	// Create a new prompt from a langchain prompt
	promptTemplate, err := template.NewFromLangchain("lc://prompts/summarize/stuff/prompt.yaml")
	if err != nil {
		panic(err)
	}

	// Format the prompt with some inputs
	output, err := promptTemplate.Format(template.Inputs{
		"text": "This is a test",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(output)

}
