# 🪿 LinGoose

[![Build Status](https://github.com/henomis/lingoose/actions/workflows/test.yml/badge.svg)](https://github.com/henomis/lingoose/actions/workflows/test.yml) [![GoDoc](https://godoc.org/github.com/henomis/lingoose?status.svg)](https://godoc.org/github.com/henomis/lingoose) [![Go Report Card](https://goreportcard.com/badge/github.com/henomis/lingoose)](https://goreportcard.com/report/github.com/henomis/lingoose) [![GitHub release](https://img.shields.io/github/release/henomis/lingoose.svg)](https://github.com/henomis/lingoose/releases)

**LinGoose** is a Go framework for creating LLM (Language Learning Machine) pipelines.
> :warning: It is a work in progress, and is not yet ready for production use. **API are unstable. Do not use in production.**

# Overview
**LinGoose** aims to be a complete framework for creating LLM apps. :robot: :gear:

# Components
**LinGoose** is composed of multiple components, each one with its own purpose.
## Prompt

Please refer to the [examples directory](examples/prompt/) to see other examples.

<details>
<summary>Hello world</summary>

```go
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
```
</details>

<details>
<summary>Prompt chat template</summary>

```go
package main

import (
	"fmt"

	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/prompt/chat"
)

func main() {

	chatTemplate := chat.New(
		[]chat.PromptMessage{
			{
				Type: chat.MessageTypeSystem,
				Prompt: &prompt.Prompt{
					Input: map[string]string{
						"input_language":  "English",
						"output_language": "French",
					},
					OutputDecoder: nil,
					Template:      "Translating from {{.input_language}} to {{.output_language}}",
				},
			},
			{
				Type: chat.MessageTypeUser,
				Prompt: &prompt.Prompt{
					Input: map[string]string{
						"text": "I love programming.",
					},
					OutputDecoder: nil,
					Template:      "{{.text}}",
				},
			},
		},
	)

	messages, err := chatTemplate.ToMessages()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", messages)

}
```
</details>


# Current development status
🟢 ready | 🟡 developing | 🔴 not started

- 🟢 Prompts
    - 🟢 Templates
    - 🟢 Chat
    - 🟢 Output decoders
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