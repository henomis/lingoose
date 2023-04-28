package main

import (
	"context"
	"fmt"
	"strconv"

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

	indexSize, _ := docsVectorIndex.Size()

	if indexSize == 0 {
		loader, err := loader.NewDirectoryLoader(".", ".txt")
		if err != nil {
			panic(err)
		}

		documents, err := loader.Load()
		if err != nil {
			panic(err)
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

		docsVectorIndex.LoadFromDocuments(context.Background(), documentChunks)
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

		id, _ := strconv.Atoi(similarity.ID)

		fmt.Printf("Similarity: %f\n", similarity.Score)
		fmt.Printf("Document: %s\n", docsVectorIndex.Data[id].Document.Content)
		fmt.Println("Metadata: ", docsVectorIndex.Data[id].Document.Metadata)
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

	llmOpenAI.Completion(context.Background(), prompt1.Prompt())

}
