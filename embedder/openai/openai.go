package openaiembedder

import (
	"context"
	"os"

	"github.com/henomis/lingoose/embedder"
	embobserver "github.com/henomis/lingoose/embedder/observer"
	"github.com/henomis/lingoose/observer"
	"github.com/sashabaranov/go-openai"
)

type Model = openai.EmbeddingModel

const (
	AdaEmbeddingV2  Model = openai.AdaEmbeddingV2
	SmallEmbedding3 Model = openai.SmallEmbedding3
	LargeEmbedding3 Model = openai.LargeEmbedding3
)

type OpenAIEmbedder struct {
	openAIClient    *openai.Client
	model           Model
	Name            string
	observer        embobserver.EmbeddingObserver
	observerTraceID string
}

func New(model Model) *OpenAIEmbedder {
	openAIKey := os.Getenv("OPENAI_API_KEY")

	return &OpenAIEmbedder{
		openAIClient: openai.NewClient(openAIKey),
		model:        model,
		Name:         "openai",
	}
}

// WithClient sets the OpenAI client to use for the embedder
func (o *OpenAIEmbedder) WithClient(client *openai.Client) *OpenAIEmbedder {
	o.openAIClient = client
	return o
}

func (o *OpenAIEmbedder) WithObserver(observer embobserver.EmbeddingObserver, traceID string) *OpenAIEmbedder {
	o.observer = observer
	o.observerTraceID = traceID
	return o
}

// Embed returns the embeddings for the given texts
func (o *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	var observerEmbedding *observer.Embedding
	var err error

	if o.observer != nil {
		observerEmbedding, err = embobserver.StartObserveEmbedding(
			o.observer,
			o.Name,
			string(o.model),
			nil,
			o.observerTraceID,
			observer.ContextValueParentID(ctx),
			texts,
		)
		if err != nil {
			return nil, err
		}
	}

	embeddings, err := o.openAICreateEmebeddings(ctx, texts)
	if err != nil {
		return nil, err
	}

	if o.observer != nil {
		err = embobserver.StopObserveEmbedding(
			o.observer,
			observerEmbedding,
			embeddings,
		)
		if err != nil {
			return nil, err
		}
	}

	return embeddings, nil
}

func (o *OpenAIEmbedder) openAICreateEmebeddings(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	resp, err := o.openAIClient.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Input: texts,
			Model: o.model,
		},
	)
	if err != nil {
		return nil, err
	}

	var embeddings []embedder.Embedding

	for _, obj := range resp.Data {
		embeddings = append(embeddings, float32ToFloat64(obj.Embedding))
	}

	return embeddings, nil
}

func float32ToFloat64(slice []float32) []float64 {
	newSlice := make([]float64, len(slice))
	for i, v := range slice {
		newSlice[i] = float64(v)
	}
	return newSlice
}
