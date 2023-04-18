# 🪿 LinGoose

[![Build Status](https://github.com/henomis/lingoose/actions/workflows/test.yml/badge.svg)](https://github.com/henomis/lingoose/actions/workflows/test.yml) [![GoDoc](https://godoc.org/github.com/henomis/lingoose?status.svg)](https://godoc.org/github.com/henomis/lingoose) [![Go Report Card](https://goreportcard.com/badge/github.com/henomis/lingoose)](https://goreportcard.com/report/github.com/henomis/lingoose) [![GitHub release](https://img.shields.io/github/release/henomis/lingoose.svg)](https://github.com/henomis/lingoose/releases)

**LinGoose** is a Go framework for creating LLM (Language Learning Machine) pipelines.
> :warning: It is a work in progress, and is not yet ready for production use. **API are unstable. Do not use in production.**

# Overview
**LinGoose** aims to be a complete framework for creating LLM apps. :robot: :gear:

# Modules
**LinGoose** is composed of multiple modules, each one with its own purpose.
## Prompt

Please refer to the [examples directory](examples/prompt/) to see other examples.
<details>
<summary>Simple prompt template</summary>

```go
package main

import (
	"fmt"

	"github.com/henomis/lingoose/prompt/template"
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
```
</details>

<details>
<summary>Prompt template with examples</summary>

```go
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
```
</details>

<details>
<summary>Prompt chat template</summary>

```go
package main

import (
	"fmt"

	"github.com/henomis/lingoose/prompt/chat"
	"github.com/henomis/lingoose/prompt/template"
)

func main() {

	chatTemplate := chat.New(
		[]chat.MessageTemplate{
			{
				Type: chat.MessageTypeSystem,
				Template: template.New(
					[]string{"input_language", "output_language"},
					[]string{},
					"You are a helpful assistant that translates {{.input_language}} to {{.output_language}}.",
					nil,
				),
			},
			{
				Type: chat.MessageTypeUser,
				Template: template.New(
					[]string{"text"},
					[]string{},
					"{{.text}}",
					nil,
				),
			},
		},
	)

	messages, err := chatTemplate.ToMessages(
		template.Inputs{
			"input_language":  "English",
			"output_language": "French",
			"text":            "I love programming.",
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", messages)

}
```
</details>

### Current development status
🟢 ready | 🟡 developing | 🔴 not started

- 🟢 Prompts
    - 🟢 Prompt Templates
    - 🟢 Prompt Examples
    - 🟢 Prompt Serialization
    - 🟢 Chat Prompt Templates
    - 🔴 Output parser
    - 🔴 Output variables
- 🟡 LLM
- 🔴 Pipelines
- 🔴 Agents

# Installation
Be sure to have a working Go environment, then run the following command:

```shell
go get github.com/henomis/lingoose
```

# License
© Simone Vellei, 2023~`time.Now()`
Released under the [MIT License](LICENSE)