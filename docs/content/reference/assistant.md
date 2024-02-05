---
title: "AI Assistant"
description:
linkTitle: "Assistant"
menu: { main: { parent: 'reference', weight: -90 } }
---

An AI Assistant is a conversational component that can understand and respond to natural language. It can be used to automate tasks, answer questions, and provide information. LinGoose offers an AI Assistant that is built on top of the `Thread`, `LLM` and `RAG` components. It uses the RAG model to retrieve relevant documents and then uses a language model to generate responses based on the retrieved documents. 

## Using AI Assistant

LinGoose assistant can optionally be configured with a RAG model and a language model. 

```go
myAssistant := assistant.New(
    openai.New().WithTemperature(0),
).WithRAG(myRAG).WithThread(
    thread.New().AddMessages(
        thread.NewUserMessage().AddContent(
            thread.NewTextContent("what is the purpose of NATO?"),
        ),
    ),
)


err = myAssistant.Run(context.Background())
if err != nil {
    panic(err)
}

fmt.Println(myAssistant.Thread())
```

We can define the LinGoose `Assistant` as a `Thread` runner with an optional `RAG` component that will help to produce the response. 