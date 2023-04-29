package main

import (
	"context"
	"fmt"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/loader"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/textsplitter"
)

func main() {

	openaiEmbedder, err := openaiembedder.New(openaiembedder.AdaEmbeddingV2)
	if err != nil {
		panic(err)
	}

	docsVectorIndex, err := index.NewSimpleVectorIndex("docs", ".", openaiEmbedder)
	if err != nil {
		panic(err)
	}

	indexIsEmpty, _ := docsVectorIndex.IsEmpty()

	if indexIsEmpty {
		ingestData(openaiEmbedder)
	}

	query := "What is the purpose of the NATO Alliance?"
	topk := 3
	similarities, err := docsVectorIndex.SimilaritySearch(
		context.Background(),
		query,
		&topk,
	)
	if err != nil {
		panic(err)
	}

	for _, similarity := range similarities {
		fmt.Printf("Similarity: %f\n", similarity.Score)
		fmt.Printf("Document: %s\n", similarity.Document.Content)
		fmt.Println("Metadata: ", similarity.Document.Metadata)
		fmt.Println("----------")
	}

	llmOpenAI, err := openai.New(openai.GPT3TextDavinci003, openai.DefaultOpenAITemperature, openai.DefaultOpenAIMaxTokens, true)
	if err != nil {
		panic(err)
	}

	prompt1, err := prompt.NewPromptTemplate(
		"Based on the following context answer to the question.\n\nContext:\n{{.context}}\n\nQuestion: {{.query}}",
		map[string]string{
			"query":   query,
			"context": similarities[0].Document.Content,
		},
	)
	if err != nil {
		panic(err)
	}

	err = prompt1.Format(nil)
	if err != nil {
		panic(err)
	}

	_, err = llmOpenAI.Completion(context.Background(), prompt1.Prompt())

	if err != nil {
		panic(err)
	}
}

func ingestData(openaiEmbedder index.Embedder) error {
	docsVectorIndex, err := index.NewSimpleVectorIndex("docs", ".", openaiEmbedder)
	if err != nil {
		return err
	}

	loader, err := loader.NewDirectoryLoader(".", ".txt")
	if err != nil {
		return err
	}

	documents, err := loader.Load()
	if err != nil {
		return err
	}

	textSplitter := textsplitter.NewRecursiveCharacterTextSplitter(1000, 20, nil, nil)

	documentChunks := textSplitter.SplitDocuments(documents)

	for _, doc := range documentChunks {
		fmt.Println(doc.Content)
		fmt.Println("----------")
		fmt.Println(doc.Metadata)
		fmt.Println("----------")
		fmt.Println()

	}

	err = docsVectorIndex.LoadFromDocuments(context.Background(), documentChunks)
	if err != nil {
		return err
	}

	return nil
}
