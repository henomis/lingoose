package ollamaembedder

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/henomis/restclientgo"

	"github.com/rsest/lingoose/embedder"
	embobserver "github.com/rsest/lingoose/embedder/observer"
)

const (
	defaultModel    = "llama2"
	defaultEndpoint = "http://localhost:11434/api"
)

type OllamaEmbedError struct {
	Err error
}

func (e *OllamaEmbedError) Error() string {
	return fmt.Sprintf("Error embedding text: %v", e.Err)
}

type Embedder struct {
	model      string
	restClient *restclientgo.RestClient
	name       string
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

// Embed returns the embeddings for the given texts
func (e *Embedder) Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	observerEmbedding, err := embobserver.StartObserveEmbedding(
		ctx,
		e.name,
		e.model,
		nil,

		texts,
	)
	if err != nil {
		return nil, err
	}

	embeddings := make([]embedder.Embedding, len(texts))
	for i, text := range texts {
		embedding, errEmbedd := e.embed(ctx, text)
		if errEmbedd != nil {
			return nil, errEmbedd
		}
		embeddings[i] = embedding
	}

	err = embobserver.StopObserveEmbedding(
		ctx,
		observerEmbedding,
		embeddings,
	)
	if err != nil {
		return nil, err
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

	if resp.HTTPStatusCode >= http.StatusBadRequest {
		return nil, &OllamaEmbedError{
			Err: errors.New(string(resp.RawBody)),
		}
	}

	return resp.Embedding, nil
}
