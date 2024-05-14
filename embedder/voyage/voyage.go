package voyageembedder

import (
	"context"
	"net/http"
	"os"

	"github.com/henomis/restclientgo"

	"github.com/henomis/lingoose/embedder"
	embobserver "github.com/henomis/lingoose/embedder/observer"
	"github.com/henomis/lingoose/observer"
)

const (
	defaultModel    = "voyage-2"
	defaultEndpoint = "https://api.voyageai.com/v1"
)

type Embedder struct {
	model           string
	restClient      *restclientgo.RestClient
	name            string
	observer        embobserver.EmbeddingObserver
	observerTraceID string
}

func New() *Embedder {
	apiKey := os.Getenv("VOYAGE_API_KEY")

	return &Embedder{
		restClient: restclientgo.New(defaultEndpoint).WithRequestModifier(
			func(req *http.Request) *http.Request {
				req.Header.Set("Authorization", "Bearer "+apiKey)
				return req
			}),
		model: defaultModel,
		name:  "voyage",
	}
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
			e.model,
			nil,
			e.observerTraceID,
			observer.ContextValueParentID(ctx),
			texts,
		)
		if err != nil {
			return nil, err
		}
	}

	embeddings, err := e.embed(ctx, texts)
	if err != nil {
		return nil, err
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
func (e *Embedder) embed(ctx context.Context, text []string) ([]embedder.Embedding, error) {
	resp := &response{}
	err := e.restClient.Post(
		ctx,
		&request{
			Input: text,
			Model: e.model,
		},
		resp,
	)
	if err != nil {
		return nil, err
	}

	embeddings := make([]embedder.Embedding, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}
