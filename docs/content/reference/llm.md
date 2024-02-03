---
title: "Large Language Models"
description:
linkTitle: "LLM"
menu: { main: { parent: 'reference', weight: -99 } }
---


Large Language Models (LLMs) are a type of artificial intelligence model that are trained on a vast amount of text data. They are capable of understanding and generating human-like text, making them incredibly versatile for a wide range of applications.

Thanks to LLM API providers, we now have access to a variety of solutions. LinGoose defines LLMs as an interface  exposing a set of methods to interact with the model. This allows developers to use the same code to interact with different LLMs, regardless of the underlying implementation.

LinGoose supports the following LLM providers API:
- [OpenAI](https://openai.com)
- [Cohere](https://cohere.ai)
- [Huggingface](https://huggingface.co)
- [Ollama](https://ollama.ai)
- [LocalAI](https://localai.io/) (_via OpenAI API compatibility_)

## Using LLMs

To use an LLM, you need to create an instance of your preferrede LLM provider. Here we show how to create an instance of an LLM using the OpenAI API:

```go
openaiLLM := openai.New()
```

A LinGoose LLM implementation may have different methods to configure the behavior of the model, such as setting the temperature, maximum tokens, and other parameters. Here is an example of how to set the temperature, maximum tokens, and model for an OpenAI LLM:

```go
openaiLLM := openai.New().
    WithTemperature(0.5).
    WithMaxTokens(1000).
    WithModel(openai.GPT4)
```

By default LinGoose LLM will use the API key from the related environment variable (e.g. `OPENAI_API_KEY` for OpenAI).

## Generating responses

To generate a response from an LLM, you need to call the `Generate` method and pass a thread containing the user, system or assistant messages. Here is an example of how to generate a response from an OpenAI LLM:

```go
myThread := thread.New().AddMessage(
    thread.NewSystemMessage().AddContent(
        thread.NewTextContent("You are a powerful AI assistant."),
    ),
).AddMessage(
    thread.NewUserMessage().AddContent(
        thread.NewTextContent("Hello, how are you?"),
    ),
)

err = openaiLLM.Generate(context.Background(), myThread)
if err != nil {
    panic(err)
}

fmt.Println(myThread)
```

Generating a response from an LLM will update the thread with the AI assistant's response. You can access the thread's history using the `Messages` field. To print the thread's history, you can use the `String` method.

## OpenAI tools call

LinGoose supports function binding to OpenAI tools. You can use the `BindFunction` method to add a tool to a the OpenAI LLM instance. Here is an example of how to add a tool call to a message:

```go
type currentWeatherInput struct {
	City string `json:"city" jsonschema:"description=City to get the weather for"`
}

func getCurrentWeather(input currentWeatherInput) string {
	// your code here
	return "The forecast for " + input.City + " is sunny."
}

...

toolChoice  := "auto"
openaiLLM := openai.New().WithToolChoice(&toolChoice)
err := openaiLLM.BindFunction(
    getCurrentWeather,
    "getCurrentWeather",
    "use this function to get the current weather for a city",
)
if err != nil {
    panic(err)
}

myThread := thread.New().AddMessage(
    thread.NewUserMessage().AddContent(
        thread.NewTextContent("I want to ride my bicycle here in Rome, but I don't know the weather."),
    ),
)

err = openaiLLM.Generate(context.Background(), myThread)
if err != nil {
    panic(err)
}

if myThread.LastMessage().Role == thread.RoleTool {

    myThread.LastMessage()

    // last message is a tool call, enrich the thread with the assistant response
    // based on the tool call result
    err = openaiLLM.Generate(context.Background(), myThread)
    if err != nil {
        panic(err)
    }
}

fmt.Println(myThread)
```

LinGoose allows you to bind a function describing its scope and input's schema. The function will be called by the OpenAI LLM automatically depending on the user's input. Here we force the tool choice to be "auto" to let OpenAI decide which tool to use. If, after an LLM generation, the last message is a tool call, you can enrich the thread with a new LLM generation based on the tool call result.


## Private LLMs
If you want to run your model or use a private LLM provider, you have many options.

### Implementing a custom LLM provider
LinGoose LLM is an interface that can be implemented by any LLM provider. You can create your own LLM by satisfying the LLM interface.

### Using a local LLM
LinGoose allows you to use to use a local LLM. You can use either LocalAI or Ollama, which are both local LLM providers.
- **LocalAI** is fully compatible with OpenAI API, so you can use it as an OpenAI LLM with a custom client configuration (`WithClient(client *openai.Client)`) pointing to your local LLM endpoint.
- **Ollama** is a local LLM provider that can be used with various LLMs, such as `llama`, `mistral`, and others.

Here is an example of how to use Ollama as LLM:

```go
myThread := thread.New()
myThread.AddMessage(thread.NewUserMessage().AddContent(
    thread.NewTextContent("How are you?"),
))

err := ollama.New().WithEndpoint("http://localhost:11434/api").WithModel("mistral").
    WithStream(func(s string) {
        fmt.Print(s)
    }).Generate(context.Background(), myThread)
if err != nil {
    panic(err)
}

fmt.Println(myThread)
```