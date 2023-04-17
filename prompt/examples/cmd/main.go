package main

import (
	"fmt"

	"github.com/henomis/lingopipe/prompt"
)

func main() {

	// Create a new prompt from a langchain prompt
	myprompt, err := prompt.NewFromLangchain("lc://prompts/summarize/stuff/prompt.yaml")
	if err != nil {
		panic(err)
	}

	// Format the prompt with some inputs
	output, err := myprompt.Format(prompt.Inputs{
		"text": "This is a test",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Output:")
	fmt.Println(output)
	fmt.Println("-------")
	fmt.Println()

	myprompt = prompt.New(
		[]string{"name"},
		[]string{},
		"Hello {{.name}}",
		nil,
	)

	// Format the prompt with some inputs
	output, err = myprompt.Format(prompt.Inputs{
		"name": "World",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Output:")
	fmt.Println(output)
	fmt.Println("-------")
	fmt.Println()

	promptExamples := prompt.PromptExamples{
		Examples: []prompt.Example{
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
		PromptTemplate: prompt.New(
			[]string{"question", "answer"},
			[]string{},
			"Question: {{.question}}\nAnswer: {{.answer}}",
			nil,
		),
	}

	myprompt, err = prompt.NewWithExamples(
		[]string{"input"},
		[]string{},
		promptExamples,
	)
	if err != nil {
		panic(err)
	}

	// Format the prompt with some inputs
	output, err = myprompt.Format(prompt.Inputs{
		"input": "World is a color?",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Output:")
	fmt.Println(output)
	fmt.Println("-------")
	fmt.Println()

	// partials
	myprompt = prompt.New(
		[]string{"foo", "bar"},
		[]string{},
		"{{.foo}}{{.bar}}",
		&prompt.Inputs{
			"bar": "baz",
		},
	)

	// Format the prompt with some inputs
	output, err = myprompt.Format(prompt.Inputs{
		"foo": "foo",
		"bar": "bar",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Output:")
	fmt.Println(output)
	fmt.Println("-------")
	fmt.Println()

	// partials 2
	myprompt = prompt.New(
		[]string{"foo", "bar"},
		[]string{},
		"{{.foo}}{{.bar}}",
		nil,
	)

	myprompt.SetPartials(&prompt.Inputs{
		"bar": "baz",
	})

	// Format the prompt with some inputs
	output, err = myprompt.Format(prompt.Inputs{
		"foo": "foo",
		"bar": "bar",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Output:")
	fmt.Println(output)
	fmt.Println("-------")
	fmt.Println()
}
