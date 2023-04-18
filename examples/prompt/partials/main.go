package main

import (
	"fmt"

	"github.com/henomis/lingoose/prompt/template"
)

func main() {
	promptTemplate := template.New(
		[]string{"foo", "bar"},
		[]string{},
		"{{.foo}}{{.bar}}",
		nil,
	)

	output, err := promptTemplate.Format(template.Inputs{
		"foo": "foo",
		"bar": "bar",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(output)

	promptTemplate.SetPartials(template.Inputs{
		"bar": "baz",
	})

	output, err = promptTemplate.Format(template.Inputs{
		"foo": "foo",
		"bar": "bar",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(output)

	promptTemplate = template.New(
		[]string{"foo", "bar"},
		[]string{},
		"{{.foo}}{{.bar}}",
		template.Inputs{
			"bar": "baz",
		},
	)

	output, err = promptTemplate.Format(template.Inputs{
		"foo": "foo",
		"bar": "bar",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(output)

}
