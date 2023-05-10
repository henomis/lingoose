package main

import (
	"context"
	"fmt"
	"os"

	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/loader"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/textsplitter"
	pineconego "github.com/henomis/pinecone-go"
	pineconerequest "github.com/henomis/pinecone-go/request"
	pineconeresponse "github.com/henomis/pinecone-go/response"
)

var pineconeClient *pineconego.PineconeGo

func main() {

	openaiEmbedder := openaiembedder.New(openaiembedder.AdaEmbeddingV2)

	pineconeApiKey := os.Getenv("PINECONE_API_KEY")
	if pineconeApiKey == "" {
		panic("PINECONE_API_KEY is not set")
	}

	pineconeEnvironment := os.Getenv("PINECONE_ENVIRONMENT")
	if pineconeEnvironment == "" {
		panic("PINECONE_ENVIRONMENT is not set")
	}

	pineconeClient = pineconego.New(pineconeEnvironment, pineconeApiKey)

	projectID, err := getProjectID(pineconeEnvironment, pineconeApiKey)
	if err != nil {
		panic(err)
	}

	pineconeIndex := index.NewPinecone(
		index.PineconeOptions{
			IndexName:      "test",
			ProjectID:      projectID,
			Namespace:      "test-namespace",
			IncludeContent: true,
		},
		openaiEmbedder,
	)

	indexIsEmpty, err := pineconeIndex.IsEmpty(context.Background())
	if err != nil {
		panic(err)
	}

	if indexIsEmpty {
		err = ingestData(projectID, openaiEmbedder)
		if err != nil {
			panic(err)
		}
	}

	query := "What is the purpose of the NATO Alliance?"
	topk := 3
	similarities, err := pineconeIndex.SimilaritySearch(
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
		fmt.Println("ID: ", similarity.ID)
		fmt.Println("----------")
	}

	llmOpenAI := openai.NewCompletion()

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

	_, err = llmOpenAI.Completion(context.Background(), prompt1.String())
	if err != nil {
		panic(err)
	}

}

func getProjectID(pineconeEnvironment, pineconeApiKey string) (string, error) {
	pineconeClient = pineconego.New(pineconeEnvironment, pineconeApiKey)

	whoamiReq := &pineconerequest.Whoami{}
	whoamiResp := &pineconeresponse.Whoami{}

	err := pineconeClient.Whoami(context.Background(), whoamiReq, whoamiResp)
	if err != nil {
		return "", err
	}

	return whoamiResp.ProjectID, nil
}

func ingestData(projectID string, openaiEmbedder index.Embedder) error {

	pineconeIndex := index.NewPinecone(
		index.PineconeOptions{
			IndexName:      "test",
			ProjectID:      projectID,
			Namespace:      "test-namespace",
			IncludeContent: true,
		},
		openaiEmbedder,
	)

	documents, err := loader.NewDirectoryLoader(".", ".txt").Load()
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

	return pineconeIndex.LoadFromDocuments(context.Background(), documentChunks)

}
