package ollamaembedder

import (
	"context"

	"github.com/henomis/restclientgo"

	"github.com/henomis/lingoose/embedder"
	embobserver "github.com/henomis/lingoose/embedder/observer"
	"github.com/henomis/lingoose/observer"
)

const (
	defaultModel    = "llama2"
	defaultEndpoint = "http://localhost:11434/api"
)

type Embedder struct {
	model           string
	restClient      *restclientgo.RestClient
	name            string
	observer        embobserver.EmbeddingObserver
	observerTraceID string
}

func New() *Embedder {
	return &Embedder{
		restClient: restclientgo.New(defaultEndpoint),
		model:      defaultModel,
		name:       "ollama",
	}
}

func (e *Embedder) WithEndpoint(endpoint string) *Embedder {
	e.restClient.SetEndpoint(endpoint)
	return e
}

func (e *Embedder) WithModel(model string) *Embedder {
	e.model = model
	return e
}

func (e *Embedder) WithObserver(observer embobserver.EmbeddingObserver, traceID string) *Embedder {
	e.observer = observer
	e.observerTraceID = traceID
	return e
}

// Embed returns the embeddings for the given texts
func (e *Embedder) Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	var observerEmbedding *observer.Embedding
	var err error

	if e.observer != nil {
		observerEmbedding, err = embobserver.StartObserveEmbedding(
			e.observer,
			e.name,
			string(e.model),
			nil,
			e.observerTraceID,
			observer.ContextValueParentID(ctx),
			texts,
		)
		if err != nil {
			return nil, err
		}
	}

	embeddings := make([]embedder.Embedding, len(texts))
	for i, text := range texts {
		embedding, err := e.embed(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}

	if e.observer != nil {
		err = embobserver.StopObserveEmbedding(
			e.observer,
			observerEmbedding,
			embeddings,
		)
		if err != nil {
			return nil, err
		}
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
