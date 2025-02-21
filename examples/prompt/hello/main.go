package main

import (
	"fmt"

	"github.com/rsest/lingoose/legacy/prompt"
	"github.com/rsest/lingoose/types"
)

type Inputs struct {
	Name string `json:"name"`
}

func main() {

	var input Inputs
	input.Name = "world"

	promptTemplate := prompt.NewPromptTemplate(
		"Hello {{.Name}}. How are {{.you}}?").WithInputs(
		input,
	)

	err := promptTemplate.Format(types.M{"you": "you"})
	if err != nil {
		panic(err)
	}

	fmt.Println(promptTemplate)

}
