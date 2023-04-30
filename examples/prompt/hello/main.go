package main

import (
	"fmt"

	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
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

	err = promptTemplate.Format(types.M{"you": "you"})
	if err != nil {
		panic(err)
	}

	fmt.Println(promptTemplate)

}
