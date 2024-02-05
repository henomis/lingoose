---
title: "Vector storage"
description:
linkTitle: "Index"
menu: { main: { parent: 'reference', weight: -93 } }
---

Vector storage is a component of LinGoose that provides a way to store and retrieve vectors. It is used to store embeddings, which are vector representations of words, phrases, or documents in a high-dimensional space. These vectors capture semantic and contextual information about the text, allowing algorithms to understand relationships and similarities between words.

LinGoose provides the `Index` interface for working with vector storage, allowing developers to use the same code to interact with different vector storage providers, regardless of the underlying implementation. LinGoose supports the following vector storage providers:

- JsonDB (it's a simple internal JSON file storage)
- [Pinecone](https://pinecone.io)
- [Qdrant](https://qdrant.tech)
- [Redis](https://redis.io)
- [PostgreSQL](https://www.postgresql.org)
- [Milvus](https://milvus.io)

## Using Index

To use an index, you need to create an instance of your preferred index provider. Here we show how to create an instance of an index using the JsonDB provider:

```go
qdrantIndex := index.New(
    qdrantdb.New(
        qdrantdb.Options{
            CollectionName: "test",
            CreateCollection: &qdrantdb.CreateCollectionOptions{
                Dimension: 1536,
                Distance:  qdrantdb.DistanceCosine,
            },
        },
    ).WithAPIKeyAndEdpoint("", "http://localhost:6333"),
    openaiembedder.New(openaiembedder.AdaEmbeddingV2),
).WithIncludeContents(true)
```

An Index instance requires an Embedder to be passed in. The Embedder is used to convert text into vectors and perform similarity searches. In this example, we use the `openaiembedder` package to create an instance of the OpenAI Embedding service. This examples uses a local Qdrant instance. Every Index provider has its own configuration options, in this case we are creating a collection with a dimension of 1536 and using the cosine distance metric and forcing the index to include the metadata contents.

To ingest a document into the index, you can use the `LoadFromDocuments` method:

```go
err := qdrantIndex.LoadFromDocuments(context.Background(), documents)
```

To search for similar documents, you can use the `Search` method:

```go
query := "What is the purpose of the NATO Alliance?"
similarities, err := index.Query(
    context.Background(),
    query,
    indexoption.WithTopK(3),
)
if err != nil {
    panic(err)
}
```

The `Query` method returns a list of `SearchResult` objects, which contain the document ID and the similarity score. The `WithTopK` option is used to specify the number of similar documents to return.
