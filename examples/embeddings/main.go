package main

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/document"
	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/loader"
	"github.com/henomis/lingoose/textsplitter"
)

func loadDocuments() *index.SimpleVectorIndex {

	docsVectorIndex := new(index.SimpleVectorIndex)
	_, err := os.Stat("docs.json")
	if err == nil {
		docsVectorIndex.Load("docs.json")
		return docsVectorIndex
	}

	loader, err := loader.NewDirectoryLoader(".", ".txt")
	if err != nil {
		panic(err)
	}

	documents, err := loader.Load()
	if err != nil {
		panic(err)
	}

	textSplitter := textsplitter.NewRecursiveCharacterTextSplitter(1000, 0, nil, nil)

	documentChunks := textSplitter.SplitDocuments(documents)

	for _, doc := range documentChunks {
		fmt.Println(doc.Content)
		fmt.Println("----------")
		fmt.Println(doc.Metadata)
		fmt.Println("----------")
		fmt.Println()

	}

	openaiEmbedder, err := openaiembedder.New(openaiembedder.AdaEmbeddingV2)
	if err != nil {
		panic(err)
	}

	objs, err := openaiEmbedder.Embed(context.Background(), documentChunks)
	if err != nil {
		panic(err)
	}

	vector := index.NewSimpleVectorIndex(documentChunks, objs)
	vector.Save("docs.json")

	docsVectorIndex.Load("docs.json")

	return docsVectorIndex
}

func loadQuery() *index.SimpleVectorIndex {

	queryVectorIndex := new(index.SimpleVectorIndex)
	_, err := os.Stat("query.json")
	if err == nil {
		queryVectorIndex.Load("query.json")
		return queryVectorIndex
	}

	docs := []document.Document{
		{
			Content: "What is the purpose of the NATO Alliance?",
		},
	}

	embed, err := openaiembedder.New(openaiembedder.AdaEmbeddingV2)
	if err != nil {
		panic(err)
	}

	objs, err := embed.Embed(context.Background(), docs)
	if err != nil {
		panic(err)
	}

	vector := index.NewSimpleVectorIndex(docs, objs)
	vector.Save("query.json")

	queryVectorIndex.Load("query.json")

	return queryVectorIndex
}

func main() {

	vectorDocs := loadDocuments()

	vectorQuery := loadQuery()

	topk := 3

	similarities := vectorDocs.Search(vectorQuery.Data[0].Embedding, &topk)

	for _, similarity := range similarities {
		fmt.Printf("Similarity: %f\n", similarity.Score)
		fmt.Printf("Document: %s\n", vectorDocs.Data[similarity.Index].Document.Content)
		fmt.Println("Metadata: ", vectorDocs.Data[similarity.Index].Document.Metadata)
		fmt.Println("----------")
	}

}
