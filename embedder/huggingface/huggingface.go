package huggingfaceembedder

import (
	"context"
	"net/http"
	"os"

	"github.com/henomis/lingoose/embedder"
	embobserver "github.com/henomis/lingoose/embedder/observer"
	"github.com/henomis/lingoose/observer"
)

const (
	hfDefaultEmbedderModel = "sentence-transformers/all-MiniLM-L6-v2"
)

type HuggingFaceEmbedder struct {
	token           string
	model           string
	httpClient      *http.Client
	name            string
	observer        embobserver.EmbeddingObserver
	observerTraceID string
}

func New() *HuggingFaceEmbedder {
	return &HuggingFaceEmbedder{
		token:      os.Getenv("HUGGING_FACE_HUB_TOKEN"),
		model:      hfDefaultEmbedderModel,
		httpClient: http.DefaultClient,
		name:       "huggingface",
	}
}

// WithToken sets the API key to use for the embedder
func (h *HuggingFaceEmbedder) WithToken(token string) *HuggingFaceEmbedder {
	h.token = token
	return h
}

// WithModel sets the model to use for the embedder
func (h *HuggingFaceEmbedder) WithModel(model string) *HuggingFaceEmbedder {
	h.model = model
	return h
}

func (h *HuggingFaceEmbedder) WithObserver(
	observer embobserver.EmbeddingObserver,
	traceID string,
) *HuggingFaceEmbedder {
	h.observer = observer
	h.observerTraceID = traceID
	return h
}

// WithHTTPClient sets the http client to use for the LLM
func (h *HuggingFaceEmbedder) WithHTTPClient(httpClient *http.Client) *HuggingFaceEmbedder {
	h.httpClient = httpClient
	return h
}

// Embed returns the embeddings for the given texts
func (h *HuggingFaceEmbedder) Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	var observerEmbedding *observer.Embedding
	var err error

	if h.observer != nil {
		observerEmbedding, err = embobserver.StartObserveEmbedding(
			h.observer,
			h.name,
			h.model,
			nil,
			h.observerTraceID,
			observer.ContextValueParentID(ctx),
			texts,
		)
		if err != nil {
			return nil, err
		}
	}

	embeddings, err := h.featureExtraction(ctx, texts)
	if err != nil {
		return nil, err
	}

	if h.observer != nil {
		err = embobserver.StopObserveEmbedding(
			h.observer,
			observerEmbedding,
			embeddings,
		)
		if err != nil {
			return nil, err
		}
	}

	return embeddings, nil
}
