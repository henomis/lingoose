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

	output, err := promptTemplate.Format(template.Inputs{
		"name": "World",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(output)

}
