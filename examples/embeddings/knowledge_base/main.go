package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/chat"
	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/loader"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/textsplitter"
	"github.com/henomis/lingoose/types"
)

const (
	KB = "https://en.wikipedia.org/wiki/World_War_II"
)

func main() {

	openaiEmbedder, err := openaiembedder.New(openaiembedder.AdaEmbeddingV2)
	if err != nil {
		panic(err)
	}

	docsVectorIndex, err := index.NewSimpleVectorIndex("db", ".", openaiEmbedder)
	if err != nil {
		panic(err)
	}

	indexIsEmpty, _ := docsVectorIndex.IsEmpty()

	if indexIsEmpty {
		err = ingestData(openaiEmbedder)
		if err != nil {
			panic(err)
		}
	}

	llmOpenAI, err := openai.New(
		openai.GPT3Dot5Turbo,
		openai.DefaultOpenAITemperature,
		openai.DefaultOpenAIMaxTokens,
		false,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Enter a query to search the knowledge base. Type 'quit' to exit.")
	query := ""
	for query != "quit" {

		fmt.Printf("> ")
		reader := bufio.NewReader(os.Stdin)
		query, _ := reader.ReadString('\n')

		if query == "quit" {
			break
		}

		topk := 3
		similarities, err := docsVectorIndex.SimilaritySearch(context.Background(), query, &topk)
		if err != nil {
			panic(err)
		}

		content := ""

		for _, similarity := range similarities {
			fmt.Printf("Similarity: %f\n", similarity.Score)
			fmt.Printf("Document: %s\n", similarity.Document.Content)
			fmt.Println("Metadata: ", similarity.Document.Metadata)
			fmt.Println("----------")
			content += similarity.Document.Content + "\n"
		}

		systemPrompt := prompt.New("You are an helpful assistant. Answer to the questions using only " +
			"the provided context. Don't add any information that is not in the context. " +
			"If you don't know the answer, just say 'I don't know'.",
		)
		userPrompt, err := prompt.NewPromptTemplate(
			"Based on the following context answer to the question.\n\nContext:\n{{.context}}\n\nQuestion: {{.query}}",
			types.M{
				"query":   query,
				"context": content,
			},
		)
		if err != nil {
			panic(err)
		}

		chat := chat.New(
			chat.PromptMessage{
				Type:   chat.MessageTypeSystem,
				Prompt: systemPrompt,
			},
			chat.PromptMessage{
				Type:   chat.MessageTypeUser,
				Prompt: userPrompt,
			},
		)

		response, err := llmOpenAI.Chat(context.Background(), chat)
		if err != nil {
			panic(err)
		}

		fmt.Println(response)

	}

}

func ingestData(openaiEmbedder index.Embedder) error {

	fmt.Printf("Learning Knowledge Base...")

	docsVectorIndex, err := index.NewSimpleVectorIndex("db", ".", openaiEmbedder)
	if err != nil {
		return err
	}

	loader, err := loader.NewPDFToTextLoader("/usr/bin/pdftotext", "./kb")
	if err != nil {
		return err
	}

	documents, err := loader.Load()
	if err != nil {
		return err
	}

	textSplitter := textsplitter.NewRecursiveCharacterTextSplitter(2000, 200, nil, nil)

	documentChunks := textSplitter.SplitDocuments(documents)

	err = docsVectorIndex.LoadFromDocuments(context.Background(), documentChunks)
	if err != nil {
		return err
	}

	fmt.Printf("Done\n")

	return nil
}