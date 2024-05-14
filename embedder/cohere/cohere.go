package cohereembedder

import (
	"context"
	"os"

	coherego "github.com/henomis/cohere-go"
	"github.com/henomis/cohere-go/model"
	"github.com/henomis/cohere-go/request"
	"github.com/henomis/cohere-go/response"

	"github.com/henomis/lingoose/embedder"
	embobserver "github.com/henomis/lingoose/embedder/observer"
	"github.com/henomis/lingoose/observer"
)

type EmbedderModel = model.EmbedModel

const (
	defaultEmbedderModel EmbedderModel = model.EmbedModelEnglishV20

	EmbedderModelEnglishV20           EmbedderModel = model.EmbedModelEnglishV20
	EmbedderModelEnglishLightV20      EmbedderModel = model.EmbedModelEnglishLightV20
	EmbedderModelMultilingualV20      EmbedderModel = model.EmbedModelMultilingualV20
	EmbedderModelEnglishV30           EmbedderModel = model.EmbedModelEnglishV30
	EmbedderModelEnglishLightV30      EmbedderModel = model.EmbedModelEnglishLightV30
	EmbedderModelMultilingualV30      EmbedderModel = model.EmbedModelMultilingualV30
	EmbedderModelMultilingualLightV30 EmbedderModel = model.EmbedModelMultilingualLightV30
)

var EmbedderModelsSize = map[EmbedderModel]int{
	EmbedderModelEnglishV30:           1024,
	EmbedderModelEnglishLightV30:      384,
	EmbedderModelEnglishV20:           4096,
	EmbedderModelEnglishLightV20:      1024,
	EmbedderModelMultilingualV30:      1024,
	EmbedderModelMultilingualLightV30: 384,
	EmbedderModelMultilingualV20:      768,
}

type Embedder struct {
	model           EmbedderModel
	client          *coherego.Client
	name            string
	observer        embobserver.EmbeddingObserver
	observerTraceID string
}

func New() *Embedder {
	return &Embedder{
		client: coherego.New(os.Getenv("COHERE_API_KEY")),
		model:  defaultEmbedderModel,
		name:   "cohere",
	}
}

// WithAPIKey sets the API key to use for the embedder
func (e *Embedder) WithAPIKey(apiKey string) *Embedder {
	e.client = coherego.New(apiKey)
	return e
}

// WithModel sets the model to use for the embedder
func (e *Embedder) WithModel(model EmbedderModel) *Embedder {
	e.model = model
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

func (e *Embedder) embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	resp := &response.Embed{}
	err := e.client.Embed(
		ctx,
		&request.Embed{
			Texts: texts,
			Model: e.model,
		},
		resp,
	)
	if err != nil {
		return nil, err
	}

	embeddings := make([]embedder.Embedding, len(resp.Embeddings))

	for i, embedding := range resp.Embeddings {
		embeddings[i] = embedding
	}
	return embeddings, nil
}
