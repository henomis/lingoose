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

	openaiEmbedder := openaiembedder.New(openaiembedder.AdaEmbeddingV2)

	pineconeIndex := index.NewPinecone(
		index.PineconeOptions{
			IndexName:      "test",
			Namespace:      "test-namespace",
			IncludeContent: true,
			CreateIndex: &index.PineconeCreateIndexOptions{
				Dimension: 1536,
				Replicas:  1,
				Metric:    "cosine",
				PodType:   "p1.x1",
			},
		},
		openaiEmbedder,
	)

	indexIsEmpty, err := pineconeIndex.IsEmpty(context.Background())
	if err != nil {
		panic(err)
	}

	if indexIsEmpty {
		err = ingestData(pineconeIndex)
		if err != nil {
			panic(err)
		}
	}

	query := "What is the purpose of the NATO Alliance?"
	similarities, err := pineconeIndex.SimilaritySearch(
		context.Background(),
		query,
		index.WithTopK(3),
	)
	if err != nil {
		panic(err)
	}

	content := ""
	for _, similarity := range similarities {
		fmt.Printf("Similarity: %f\n", similarity.Score)
		fmt.Printf("Document: %s\n", similarity.Document.Content)
		fmt.Println("Metadata: ", similarity.Document.Metadata)
		fmt.Println("ID: ", similarity.ID)
		fmt.Println("----------")
		content += similarity.Document.Content + "\n"
	}

	llmOpenAI := openai.NewCompletion().WithVerbose(true)

	prompt1, err := prompt.NewPromptTemplate(
		"Based on the following context answer to the question.\n\nContext:\n{{.context}}\n\nQuestion: {{.query}}",
		map[string]string{
			"query":   query,
			"context": content,
		},
	)
	if err != nil {
		panic(err)
	}

	err = prompt1.Format(nil)
	if err != nil {
		panic(err)
	}

	_, err = llmOpenAI.Completion(context.Background(), prompt1.String())
	if err != nil {
		panic(err)
	}

}

func ingestData(pineconeIndex *index.Pinecone) error {

	documents, err := loader.NewDirectoryLoader(".", ".txt").Load()
	if err != nil {
		return err
	}

	textSplitter := textsplitter.NewRecursiveCharacterTextSplitter(1000, 20)

	documentChunks := textSplitter.SplitDocuments(documents)

	for _, doc := range documentChunks {
		fmt.Println(doc.Content)
		fmt.Println("----------")
		fmt.Println(doc.Metadata)
		fmt.Println("----------")
		fmt.Println()

	}

	return pineconeIndex.LoadFromDocuments(context.Background(), documentChunks)

}
