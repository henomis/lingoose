package main

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedding"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/loader"
	"github.com/henomis/lingoose/textsplitter"
)

type Document struct {
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

type Loader interface {
	Load() []document.Document
}

type TextSplitter interface {
	SplitDocuments([]document.Document) []document.Document
	SplitText(string) []string // ??? not sure if this is needed
}

type EmbeddingObject struct {
	Vector []float32 `json:"vector"`
	Index  int       `json:"index"`
}

type Embedder interface {
	Embed(ctx context.Context, docs []document.Document) ([]EmbeddingObject, error)
}

type Similarity struct {
	Score float32
	Index int
}

type Index interface {
	Search(embedding embedding.EmbeddingObject, topK *int) []index.Similarity
}

func loadDocuments() (*index.SimpleVector, []document.Document) {

	loader, err := loader.NewDirLoader(".", ".txt")
	if err != nil {
		panic(err)
	}

	documents, err := loader.Load()
	if err != nil {
		panic(err)
	}

	text_splitter := textsplitter.NewRecursiveCharacterTextSplitter(nil, 1000, 0, nil)

	docs := text_splitter.SplitDocuments(documents)

	for _, doc := range docs {
		fmt.Println(doc.Content)
		fmt.Println("----------")
		fmt.Println(doc.Metadata)
		fmt.Println("----------")
		fmt.Println()

	}

	embed, err := embedding.NewOpenAIEmbeddings(embedding.AdaSimilarity)
	if err != nil {
		panic(err)
	}

	vectorDocs := new(index.SimpleVector)
	_, err = os.Stat("docs.json")
	if err != nil {

		objs, err := embed.Embed(context.Background(), docs)
		if err != nil {
			panic(err)
		}

		vector := index.NewVectorIndex(objs)
		vector.Save("docs.json")
	}
	vectorDocs.Load("docs.json")

	return vectorDocs, docs
}

func loadQuery() *index.SimpleVector {

	docs := []document.Document{
		{
			Content: "What is the purpose of the NATO Alliance?",
		},
	}

	embed, err := embedding.NewOpenAIEmbeddings(embedding.AdaSimilarity)
	if err != nil {
		panic(err)
	}

	vectorDocs := new(index.SimpleVector)
	_, err = os.Stat("query.json")
	if err != nil {

		objs, err := embed.Embed(context.Background(), docs)
		if err != nil {
			panic(err)
		}

		vector := index.NewVectorIndex(objs)
		vector.Save("query.json")
	}
	vectorDocs.Load("query.json")

	return vectorDocs
}

func main() {

	vectorDocs, docs := loadDocuments()

	vectorQuery := loadQuery()

	topk := 3

	similarities := vectorDocs.Search(vectorQuery.Embeddings()[0], &topk)

	for _, similarity := range similarities {
		fmt.Printf("Similarity: %f\n", similarity.Score)
		fmt.Printf("Document: %s\n", docs[similarity.Index].Content)
		fmt.Println("----------")
	}
}
