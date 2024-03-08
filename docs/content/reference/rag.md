---
title: "Retrieval Augmented Generation"
description:
linkTitle: "RAG"
menu: { main: { parent: 'reference', weight: -91 } }
---

Retrieval Augmented Generation (RAG) is a component that combines the strengths of both retrieval-based and generation-based models. It uses a index to find relevant documents and then uses a language model to generate responses based on the retrieved documents.

LinGoose offers a RAG model that is built on top of the Index component. It uses the index to retrieve relevant documents and then uses a language model to generate responses based on the retrieved documents. The RAG model can be extended to support multiple RAG implementations.

## Using RAG

To use RAG, you need to first create an index and then use the index to retrieve relevant documents. 

```go
rag := rag.New(
    index.New(
        jsondb.New().WithPersist("index.json"),
        openaiembedder.New(openaiembedder.AdaEmbeddingV2),
    ),
).WithChunkSize(1000).WithChunkOverlap(0)
```

Here we create a new RAG model with a JSON database and an OpenAI embedder. We also set the chunk size and overlap for the index. 

```go
rag.AddDocuments(
    context.Background(),
    document.Document{
        Content: `Augusta Ada King, Countess of Lovelace (née Byron; 10 December 1815 -
            27 November 1852) was an English mathematician and writer, 
            chiefly known for her work on Charles Babbage's proposed mechanical general-purpose computer,
            the Analytical Engine. She was the first to recognise that the machine had applications beyond pure calculation.
            `,
        Metadata: types.Meta{
            "author": "Wikipedia",
        },
    },
)
```

You can add documents to the RAG knowledge using the `AddDocuments` method. You can attach to the RAG one or more Loader specifying the matching pattern for file source of the documents.

```go
rag = rag.WithLoader(regexp.MustCompile(`.*\.pdf`), loader.NewPDFToText())
```

There are default loader already attached to the RAG that you can override or extend.

- `.*\.pdf` via `loader.NewPDFToText()`
- `.*\.txt` via `loader.NewText()`
- `.*\.docx` via `loader.NewLibreOffice()`

## Fusion RAG
This is an advance RAG algorithm that uses an LLM to generate additional queries based on the original one. New queries will be used to retrieve more documents that will be reranked and used to generate the final response.

```go
fusionRAG := rag.NewFusion(
    index.New(
        jsondb.New().WithPersist("index.json"),
        openaiembedder.New(openaiembedder.AdaEmbeddingV2),
    ),
    openai.New(),
)
```

## Subdocument RAG
This is an advance RAG algorithm that ingest documents chunking them in subdocuments and attaching a summary of the parent document. This will allow the RAG to retrieve more relevant documents and generate better responses.

```go
fusionRAG := rag.NewSubDocument(
    index.New(
        jsondb.New().WithPersist("index.json"),
        openaiembedder.New(openaiembedder.AdaEmbeddingV2),
    ),
    openai.New(),
)
```