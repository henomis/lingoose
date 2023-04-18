package main

import (
	"fmt"

	"github.com/henomis/lingopipes/prompt/template"
)

func main() {

	promptTemplate := template.New(
		[]string{"name"},
		[]string{},
		"Hello {{.name}}",
		nil,
	)

	promptTemplate.Save("prompt.yaml")

	promptTemplate.Load("prompt.yaml")

	output, err := promptTemplate.Format(template.Inputs{
		"name": "World",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(output)

}
