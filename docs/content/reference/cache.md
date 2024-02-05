---
title: "Caching LLM responses"
description:
linkTitle: "Cache"
menu: { main: { parent: 'reference', weight: -92 } }
---
 
Caching LLM responses can be a good way to improve the performance of your application. This is especially true when you are using an LLM to generate responses to user messages in real-time. By caching the responses, you can avoid making repeated calls to the LLM, which can be slow and expensive.

LinGoose provides a built-in caching mechanism that you can use to cache LLM responses. The cache needs an Index to be created, and it will store the responses in memory. 

## Using the cache

To use the cache, you need to create an Index and pass it to the LLM instance. Here's an example:

```go
openAILLM := openai.New().WithCache(
    cache.New(
        index.New(
            jsondb.New().WithPersist("index.json"),
            openaiembedder.New(openaiembedder.AdaEmbeddingV2),
        ),
    ).WithTopK(3),
)
```

Here we are creating a new cache with an index that uses a JSON database to persist the data. We are also using the `openaiembedder` to perform the index embedding operations. The `WithTopK` method is used to specify the number of responses to evaluate to get the answer from the cache.

Once you have created the cache, your LLM instance will use it to store and retrieve responses. The cache will automatically store responses and retrieve them when needed.

```go
questions := []string{
    "what's github",
    "can you explain what GitHub is",
    "can you tell me more about GitHub",
    "what is the purpose of GitHub",
}

for _, question := range questions {
    t := thread.New().AddMessage(
        thread.NewUserMessage().AddContent(
            thread.NewTextContent(question),
        ),
    )

    err := llm.Generate(context.Background(), t)
    if err != nil {
        fmt.Println(err)
        continue
    }

    fmt.Println(t)
}
```

In this example, we are using the LLM to generate responses to a list of questions. The cache will store the responses and retrieve them when needed. This can help to improve the performance of your application by avoiding repeated calls to the LLM.