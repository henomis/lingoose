package ollamaembedder

import (
	"context"

	"github.com/henomis/lingoose/embedder"
	"github.com/henomis/restclientgo"
)

const (
	defaultModel    = "llama2"
	defaultEndpoint = "http://localhost:11434/api"
)

type Embedder struct {
	model      string
	restClient *restclientgo.RestClient
}

func New() *Embedder {
	return &Embedder{
		restClient: restclientgo.New(defaultEndpoint),
		model:      defaultModel,
	}
}

func (o *Embedder) WithEndpoint(endpoint string) *Embedder {
	o.restClient.SetEndpoint(endpoint)
	return o
}

func (o *Embedder) WithModel(model string) *Embedder {
	o.model = model
	return o
}

// Embed returns the embeddings for the given texts
func (e *Embedder) Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	embeddings := make([]embedder.Embedding, len(texts))
	for i, text := range texts {
		embedding, err := e.embed(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// Embed returns the embeddings for the given texts
func (e *Embedder) embed(ctx context.Context, text string) (embedder.Embedding, error) {
	resp := &response{}
	err := e.restClient.Post(
		ctx,
		&request{
			Prompt: text,
			Model:  e.model,
		},
		resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Embedding, nil
}
