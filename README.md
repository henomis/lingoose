![image](./docs/assets/img/lingoose-small.png )

# 🪿 LinGoose

[![Build Status](https://github.com/henomis/lingoose/actions/workflows/test.yml/badge.svg)](https://github.com/henomis/lingoose/actions/workflows/test.yml) [![GoDoc](https://godoc.org/github.com/henomis/lingoose?status.svg)](https://godoc.org/github.com/henomis/lingoose) [![Go Report Card](https://goreportcard.com/badge/github.com/henomis/lingoose)](https://goreportcard.com/report/github.com/henomis/lingoose) [![GitHub release](https://img.shields.io/github/release/henomis/lingoose.svg)](https://github.com/henomis/lingoose/releases)

**LinGoose** (_Lingo + Go + Goose_ 🪿) aims to be a complete Go framework for creating LLM apps. 🤖 ⚙️

> **Did you know?** A goose 🪿 fills its car 🚗 with goose-line ⛽!

here below an image from docs/assets/img/lingoose.png

# Overview
**LinGoose** is a powerful Go framework for developing Large Language Model (LLM) based applications using pipelines. It is designed to be a complete solution and provides multiple components, including Prompts, Templates, Chat, Output Decoders, LLM, Pipelines, and Memory. With **LinGoose**, you can interact with LLM AI through prompts and generate complex templates. Additionally, it includes a chat feature, allowing you to create chatbots. The Output Decoders component enables you to extract specific information from the output of the LLM, while the LLM interface allows you to send prompts to various AI, such as the ones provided by OpenAI. You can chain multiple LLM steps together using Pipelines and store the output of each step in Memory for later retrieval. **LinGoose** also includes a Document component, which is used to store text, and a Loader component, which is used to load Documents from various sources. Finally, it includes TextSplitters, which are used to split text or Documents into multiple parts, Embedders, which are used to embed text or Documents into embeddings, and Indexes, which are used to store embeddings and documents and to perform searches.

# Components
**LinGoose** is composed of multiple components, each one with its own purpose.

| Component | Package|Description |
| --- | --- | ---|
|**Prompt** | [prompt](prompt/)| Prompts are the way to interact with LLM AI. They can be simple text, or more complex templates. |
|**Prompt Template** | [prompt](prompt/)| Templates are used to generate prompts formatting a generic input using Go [text/template](https://golang.org/pkg/text/template/) package. |
|**Chat Prompt** | [chat](chat/) | Chat is the way to interact with the chat LLM AI. It can be a simple text prompt, or a more complex chatbot. |
|**Output decoders** | [decoder](decoder/) | Output decoders are used to decode the output of the LLM. They can be used to extract specific information from the output. |
|**LLMs** |[llm/openai](llm/openai/) | LLM is an interface to various AI such as the ones provided by OpenAI. It is responsible for sending the prompt to the AI and retrieving the output. |
|**Pipelines** | [pipeline](pipeline/)|Pipelines are used to chain multiple LLM steps together. |
|**Memory** | [memory/ram](memory/ram/)|Memory is used to store the output of each step. It can be used to retrieve the output of a previous step. |
|**Document** | [document](document/)|Document is used to store a text |
|**Loaders** | [loader](loader/)|Loaders are used to load Documents from various sources. |
|**TextSplitters**| [textsplitter](textsplitter/)|TextSplitters are used to split text or Documents into multiple parts. |
|**Embedders** | [embedder](embedder/)|Embedders are used to embed text or Documents into embeddings. |
|**Indexes**| [index](index/)|Indexes are used to store embeddings and documents and to perform searches. |

# Usage

Please refer to the [examples directory](examples/) to see other examples. However, here is an example of what **LinGoose** is capable of:

_Talk is cheap. Show me the [code](examples/)._ - Linus Torvalds

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/henomis/lingoose/decoder"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/memory/ram"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
)

func main() {

	llmOpenAI, err := openai.New(openai.GPT3TextDavinci003, true)
	if err != nil {
		panic(err)
	}
	cache := ram.New()

	prompt1 := prompt.New("Hello how are you?")
	pipe1 := pipeline.NewStep(
		"step1",
		llmOpenAI,
		prompt1,
		decoder.NewDefaultDecoder(),
		cache,
	)

	prompt2, _ := prompt.NewPromptTemplate(
		"Consider the following sentence.\n\nSentence:\n{{.output}}\n\n"+
			"Translate it in {{.language}}!",
		map[string]string{
			"language": "italian",
		},
	)
	pipe2 := pipeline.NewStep(
		"step2",
		llmOpenAI,
		prompt2,
		decoder.NewDefaultDecoder(),
		nil,
	)

	prompt3, _ := prompt.NewPromptTemplate(
		"Consider the following sentence.\n\nSentence:\n{{.step1.output}}"+
			"\n\nTranslate it in {{.language}}!",
		map[string]string{
			"language": "spanish",
		},
	)
	pipe3 := pipeline.NewStep(
		"step3",
		llmOpenAI,
		prompt3,
		decoder.NewDefaultDecoder(),
		cache,
	)

	pipelineSteps := pipeline.New(
		pipe1,
		pipe2,
		pipe3,
	)

	response, err := pipelineSteps.Run(nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("\n\nFinal output: %#v\n\n", response)

	fmt.Println("---Memory---")
	dump, _ := json.MarshalIndent(cache.All(), "", "  ")
	fmt.Printf("%s\n", string(dump))
}
```

Running this example will produce the following output:

```
---USER---
Hello how are you?
---AI---
I'm doing well, thank you. How about you?
---USER---
Consider the following sentence.\n\nSentence:\nI'm doing well, thank you. How about you?\n\n
                Translate it in italian!
---AI---
Sto bene, grazie. E tu come stai?
---USER---
Consider the following sentence.\n\nSentence:\nI'm doing well, thank you. How about you?
                \n\nTranslate it in spanish!
---AI---
Estoy bien, gracias. ¿Y tú


Final output: map[string]interface {}{"output":"Estoy bien, gracias. ¿Y tú"}

---Memory---
{
  "step1": {
    "output": "I'm doing well, thank you. How about you?"
  },
  "step3": {
    "output": "Estoy bien, gracias. ¿Y tú"
  }
}
```

# Installation
Be sure to have a working Go environment, then run the following command:

```shell
go get github.com/henomis/lingoose
```


# License
© Simone Vellei, 2023~`time.Now()`
Released under the [MIT License](LICENSE)
