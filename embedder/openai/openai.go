package openaiembedder

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
	"github.com/sashabaranov/go-openai"
)

type Model int

const (
	Unknown Model = iota
	AdaSimilarity
	BabbageSimilarity
	CurieSimilarity
	DavinciSimilarity
	AdaSearchDocument
	AdaSearchQuery
	BabbageSearchDocument
	BabbageSearchQuery
	CurieSearchDocument
	CurieSearchQuery
	DavinciSearchDocument
	DavinciSearchQuery
	AdaCodeSearchCode
	AdaCodeSearchText
	BabbageCodeSearchCode
	BabbageCodeSearchText
	AdaEmbeddingV2
)

type openAIEmbedder struct {
	openAIClient *openai.Client
	model        Model
}

func New(model Model) (*openAIEmbedder, error) {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	return &openAIEmbedder{
		openAIClient: openai.NewClient(openAIKey),
		model:        model,
	}, nil
}

func (t *openAIEmbedder) Embed(ctx context.Context, docs []document.Document) ([]embedder.Embedding, error) {

	input := []string{}
	for _, doc := range docs {
		input = append(input, doc.Content)
	}

	resp, err := t.openAIClient.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Input: input,
			Model: openai.EmbeddingModel(t.model),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", embedder.ErrCreateEmbedding, err)
	}

	var embeddings []embedder.Embedding

	for _, obj := range resp.Data {
		embeddings = append(embeddings, embedder.Embedding{
			Embedding: obj.Embedding,
			Index:     obj.Index,
		})
	}

	return embeddings, nil
}
