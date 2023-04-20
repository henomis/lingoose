# 🪿 LinGoose

[![Build Status](https://github.com/henomis/lingoose/actions/workflows/test.yml/badge.svg)](https://github.com/henomis/lingoose/actions/workflows/test.yml) [![GoDoc](https://godoc.org/github.com/henomis/lingoose?status.svg)](https://godoc.org/github.com/henomis/lingoose) [![Go Report Card](https://goreportcard.com/badge/github.com/henomis/lingoose)](https://goreportcard.com/report/github.com/henomis/lingoose) [![GitHub release](https://img.shields.io/github/release/henomis/lingoose.svg)](https://github.com/henomis/lingoose/releases)

**LinGoose** is a Go framework for creating LLM (Language Learning Machine) pipelines.
> :warning: It is a work in progress, and is not yet ready for production use. **API are unstable. Do not use in production.**

# Overview
**LinGoose** aims to be a complete framework for creating LLM apps. :robot: :gear:

# Components
**LinGoose** is composed of multiple components, each one with its own purpose.

| Component | Description |
| --- | --- |
|**Prompts** | Prompts are the way to interact with LLM AI. They can be simple text, or more complex templates. |
|**Templates** | Templates are used to generate prompts formatting a generic input using a text template. |
|**Chat** | Chat is the way to interact with the chat LLM AI. It can be a simple text prompt, or a more complex chatbot. |
|**Output decoders** | Output decoders are used to decode the output of the LLM. They can be used to extract specific information from the output. |
|**LLM** | LLM is an interface to various AI such as the ones provided by OpenAI. It is responsible for sending the prompt to the AI and retrieving the output. |
|**Pipelines** | Pipelines are used to chain multiple LMM steps together. |
|**Memory** | Memory is used to store the output of each step. It can be used to retrieve the output of a previous step. |

# Usage

Please refer to the [examples directory](examples/) to see other examples. However, here is an example of what **LinGoose** is capable of:

_Talk is cheap. Show me the code._ - Linus Torvalds

```go
package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/henomis/lingoose/decoder"
	"github.com/henomis/lingoose/llm"
	"github.com/henomis/lingoose/memory/ram"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	cache := ram.New()

	llm1 := &llm.LlmMock{}
	prompt1 := prompt.New("Hello how are you?")
	pipe1 := pipeline.NewStep("step1", llm1, prompt1, nil, decoder.NewDefaultDecoder(), cache)

	myout := &struct {
		First  string
		Second string
	}{}
	llm2 := &llm.JsonLllMock{}
	prompt2, _ := prompt.NewPromptTemplate(
		"It seems you are a random word generator. Your message '{{.output}}' is nonsense. Anyway I'm fine {{.value}}!",
		map[string]string{
			"value": "thanks",
		},
	)
	pipe2 := pipeline.NewStep("step2", llm2, prompt2, myout, decoder.NewJSONDecoder(), cache)

	var values []string
	prompt3, _ := prompt.NewPromptTemplate(
		"Oh! It seems you are a random JSON word generator. You generated two strings, first:'{{.First}}' and second:'{{.Second}}'. {{.value}}\n\tHowever your first message was: '{{.step1.output}}'",
		map[string]string{
			"value": "Bye!",
		},
	)
	pipe3 := pipeline.NewStep("step3", llm1, prompt3, values, decoder.NewRegExDecoder(`(\w+)\s(\w+)\s(.*)`), cache)

	pipelineSteps := pipeline.New(
		pipe1,
		pipe2,
		pipe3,
	)

	response, err := pipelineSteps.Run(nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Final output: %s\n", strings.Join(response.([]string), ", "))
	fmt.Println("---Memory---")
	dump, _ := json.MarshalIndent(cache.All(), "", "  ")
	fmt.Printf("%s\n", string(dump))
}
```

Running this example you will get the following output:

```shell
User: Hello how are you?
AI: grass television door television
User: It seems you are a random word generator. Your message 'grass television door television' is nonsense. Anyway I'm fine thanks!
AI: {"first": "wind", "second": "flower"}
User: Oh! It seems you are a random JSON word generator. You generated two strings, first:'wind' and second:'flower'. Bye!
        However your first message was: 'grass television door television'
AI: grass lake fly mountain fly
Final output: grass, lake, fly mountain fly
---Memory---
{
  "step1": {
    "output": "grass television door television"
  },
  "step2": {
    "First": "wind",
    "Second": "flower"
  },
  "step3": [
    "grass",
    "lake",
    "fly mountain fly"
  ]
}
```



# Current development status

| Component | Status |
| --- | --- |
|**Prompts** | 🟢 READY|
|**Templates** | 🟢 READY|
|**Chat** | 🟢 READY|
|**Output decoders** | 🟢 READY|
|**LLM** | 🟡 DEVELOPING|
|**Pipelines** | 🟢 READY|
|**Memory** | 🟢 READY|


# Installation
Be sure to have a working Go environment, then run the following command:

```shell
go get github.com/henomis/lingoose
```


# License
© Simone Vellei, 2023~`time.Now()`
Released under the [MIT License](LICENSE)