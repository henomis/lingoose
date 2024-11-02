---
title: "Performing tasks with Tools"
description:
linkTitle: "Tool"
menu: { main: { parent: 'reference', weight: -89 } }
---

Tools are components that can be used to perform specific tasks. They can be used to automate, answer questions, and provide information. LinGoose offers a variety of tools that can be used to perform different actions.

## Available Tools

- *Python*: It can be used to run Python code and get the output.
- *SerpApi*: It can be used to get search results from Google and other search engines.
- *Dall-e*: It can be used to generate images based on text descriptions.
- *DuckDuckGo*: It can be used to get search results from DuckDuckGo.
- *RAG*: It can be used to retrieve relevant documents based on a query.
- *LLM*: It can be used to generate text based on a prompt.
- *Shell*: It can be used to run shell commands and get the output.


## Using Tools

LinGoose tools can be used to perform specific tasks. Here is an example of using the `Python` and `serpapi` tools to get information and run Python code and get the output.

```go
auto := "auto"
myAgent := assistant.New(
    openai.New().WithModel(openai.GPT4o).WithToolChoice(&auto).WithTools(
        pythontool.New(),
        serpapitool.New(),
    ),
).WithParameters(
    assistant.Parameters{
        AssistantName:      "AI Assistant",
        AssistantIdentity:  "an helpful assistant",
        AssistantScope:     "with their questions.",
        CompanyName:        "",
        CompanyDescription: "",
    },
).WithThread(
    thread.New().AddMessages(
        thread.NewUserMessage().AddContent(
            thread.NewTextContent("calculate the average temperature in celsius degrees of New York, Rome, and Tokyo."),
        ),
    ),
).WithMaxIterations(10)

err := myAgent.Run(context.Background())
if err != nil {
    panic(err)
}
```