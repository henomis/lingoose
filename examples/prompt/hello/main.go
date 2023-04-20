package main

import (
	"fmt"

	"github.com/henomis/lingoose/prompt"
)

type Inputs struct {
	Name string `json:"name"`
}

func main() {

	var input Inputs
	input.Name = "world"

	promptTemplate, err := prompt.NewPromptTemplate(
		"Hello {{.Name}}. How are {{.you}}?",
		input,
	)
	if err != nil {
		panic(err)
	}

	err = promptTemplate.Format(map[string]interface{}{"you": "you"})
	if err != nil {
		panic(err)
	}

	fmt.Println(promptTemplate.Prompt())

}
