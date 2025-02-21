package main

import (
	"context"
	"fmt"

	"github.com/rsest/lingoose/document"
	openaiembedder "github.com/rsest/lingoose/embedder/openai"
	"github.com/rsest/lingoose/index"
	"github.com/rsest/lingoose/index/vectordb/jsondb"
	"github.com/rsest/lingoose/rag"
	"github.com/rsest/lingoose/types"
)

func main() {

	rag := rag.New(
		index.New(
			jsondb.New().WithPersist("index.json"),
			openaiembedder.New(openaiembedder.AdaEmbeddingV2),
		),
	).WithChunkSize(1000).WithChunkOverlap(0)

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

	results, err := rag.Retrieve(context.Background(), "Who was Ada Lovelace?")
	if err != nil {
		panic(err)
	}

	for _, result := range results {
		fmt.Println(result)
	}
}
