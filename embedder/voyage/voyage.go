package voyageembedder

import (
	"context"
	"net/http"
	"os"

	"github.com/henomis/restclientgo"

	"github.com/henomis/lingoose/embedder"
	embobserver "github.com/henomis/lingoose/embedder/observer"
)

const (
	defaultModel    = "voyage-2"
	defaultEndpoint = "https://api.voyageai.com/v1"
)

type Embedder struct {
	model      string
	restClient *restclientgo.RestClient
	name       string
	inputType  string
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

func (e *Embedder) WithInputType(inputType string) *Embedder {
	e.inputType = inputType
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

	embeddings, err := e.embed(ctx, texts)
	if err != nil {
		return nil, err
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
func (e *Embedder) embed(ctx context.Context, text []string) ([]embedder.Embedding, error) {
	resp := &response{}
	err := e.restClient.Post(
		ctx,
		&request{
			Input:     text,
			Model:     e.model,
			InputType: e.inputType,
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
