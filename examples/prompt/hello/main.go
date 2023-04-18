package main

import (
	"fmt"

	"github.com/henomis/lingoose/prompt"
)

type Inputs struct {
	Name string `json:"name" validate:"required"`
}

func main() {

	var input Inputs
	input.Name = "world"

	promptTemplate, err := prompt.New(
		input,
		nil,
		"Hello {{.Name}}",
	)
	if err != nil {
		panic(err)
	}

	output, err := promptTemplate.Format()
	if err != nil {
		panic(err)
	}

	fmt.Println(output)

}
