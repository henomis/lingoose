---
title: "LLM Embeddings"
description:
linkTitle: "Embedding"
menu: { main: { parent: 'reference', weight: -99 } }
---

Embeddings are a fundamental concept in machine learning and natural language processing. They are vector representations of words, phrases, or documents in a high-dimensional space. These vectors capture semantic and contextual information about the text, allowing algorithms to understand relationships and similarities between words.

Embeddings are used in various tasks such as language modeling, sentiment analysis, machine translation, and information retrieval. By representing words as vectors, models can perform operations like word similarity, word analogy, and clustering. Embeddings enable algorithms to learn from large amounts of text data and generalize their knowledge to new, unseen examples.

LinGoose provides an interface for working with embeddings, allowing developers to use the same code to interact with different embedding providers, regardless of the underlying implementation. LinGoose supports the following embedding providers:

- [OpenAI](https://openai.com)
- [Cohere](https://cohere.ai)
- [Huggingface](https://huggingface.co)
- [Ollama](https://ollama.ai)
- [LocalAI](https://localai.io/) (_via OpenAI API compatibility_)
- [Atlas Nomic](https://atlas.nomic.ai)

## Using Embeddings

To use an embedding provider, you need to create an instance of your preferred embedding provider. Here we show how to create an instance of an embedding using the OpenAI API:

```go
embeddins, err := openaiembedder.New(openaiembedder.AdaEmbeddingV2).Embed(
    context.Background(),
    []string{"What is the NATO purpose?"},
)
if err != nil {
    panic(err)
}

fmt.Printf("Embeddings: %v\n", embeddins)
```

By default LinGoose embeddings will use the API key from the related environment variable (e.g. `OPENAI_API_KEY` for OpenAI).

## Private Embeddings

If you want to run your model or use a private embedding provider, you have many options.

### Implementing a custom embedding provider
LinGoose Embedder is an interface that can be implemented by any LLM provider. You can create your own Embedder by satisfying the LLM interface.

### Using a local Embedder
LinGoose allows you to use to use a local Embedder. You can use either LocalAI or Ollama, which are both local providers.
- **LocalAI** is fully compatible with OpenAI API, so you can use it as an OpenAI Embeddings with a custom client configuration (`WithClient(client *openai.Client)`) pointing to your local endpoint.
- **Ollama** is a local Embedding provider that can be used with various models, such as `llama`, `mistral`, and others.

Here is an example of using Ollama Embedder:

```go
embeddins, err := ollamaembedder.New().
    WithEndpoint("http://localhost:11434/api").
    WithModel("mistral").
    Embed(
        context.Background(),
        []string{"What is the NATO purpose?"},
    )
if err != nil {
    panic(err)
}
```