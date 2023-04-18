package main

import (
	"fmt"

	"github.com/henomis/lingoose/prompt/example"
	"github.com/henomis/lingoose/prompt/template"
)

func main() {

	promptExamples := example.Examples{
		Examples: []example.Example{
			{
				"question": "Red is a color?",
				"answer":   "Yes",
			},
			{
				"question": "Car is a color?",
				"answer":   "No",
			},
		},
		Separator: "\n\n",
		Prefix:    "Answer to questions.",
		Suffix:    "Question: {{.input}}\nAnswer: ",
	}

	examplesTemplate := template.New(
		[]string{"question", "answer"},
		[]string{},
		"Question: {{.question}}\nAnswer: {{.answer}}",
		nil,
	)

	promptTemplate, err := template.NewWithExamples(
		[]string{"input"},
		[]string{},
		promptExamples,
		examplesTemplate,
	)
	if err != nil {
		panic(err)
	}

	output, err := promptTemplate.Format(template.Inputs{
		"input": "World is a color?",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(output)

}
