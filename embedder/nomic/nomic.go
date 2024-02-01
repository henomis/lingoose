package nomicembedder

import (
	"context"
	"net/http"
	"os"

	"github.com/henomis/lingoose/embedder"
	"github.com/henomis/restclientgo"
)

const (
	defaultEndpoint = "https://api-atlas.nomic.ai/v1"
	defaultModel    = ModelNomicEmbedTextV1
)

type Embedder struct {
	taskType   TaskType
	model      Model
	restClient *restclientgo.RestClient
}

func New() *Embedder {
	apiKey := os.Getenv("NOMIC_API_KEY")

	return &Embedder{
		restClient: restclientgo.New(defaultEndpoint).WithRequestModifier(
			func(req *http.Request) *http.Request {
				req.Header.Set("Authorization", "Bearer "+apiKey)
				return req
			},
		),
		model: defaultModel,
	}
}

func (e *Embedder) WithAPIKey(apiKey string) *Embedder {
	e.restClient = restclientgo.New(defaultEndpoint).WithRequestModifier(
		func(req *http.Request) *http.Request {
			req.Header.Set("Authorization", "Bearer "+apiKey)
			return req
		},
	)
	return e
}

func (e *Embedder) WithTaskType(taskType TaskType) *Embedder {
	e.taskType = taskType
	return e
}

func (e *Embedder) WithModel(model Model) *Embedder {
	e.model = model
	return e
}

// Embed returns the embeddings for the given texts
func (e *Embedder) Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	var resp response
	err := e.restClient.Post(
		ctx,
		&request{
			Texts:    texts,
			Model:    string(e.model),
			TaskType: e.taskType,
		},
		&resp,
	)
	if err != nil {
		return nil, err
	}

	return resp.Embeddings, nil
}
