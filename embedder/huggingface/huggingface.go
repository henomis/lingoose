package huggingfaceembedder

import (
	"context"
	"net/http"
	"os"

	"github.com/rsest/lingoose/embedder"
	embobserver "github.com/rsest/lingoose/embedder/observer"
)

const (
	hfDefaultEmbedderModel = "sentence-transformers/all-MiniLM-L6-v2"
)

type HuggingFaceEmbedder struct {
	token      string
	model      string
	httpClient *http.Client
	name       string
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

// WithHTTPClient sets the http client to use for the LLM
func (h *HuggingFaceEmbedder) WithHTTPClient(httpClient *http.Client) *HuggingFaceEmbedder {
	h.httpClient = httpClient
	return h
}

// Embed returns the embeddings for the given texts
func (h *HuggingFaceEmbedder) Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	observerEmbedding, err := embobserver.StartObserveEmbedding(
		ctx,
		h.name,
		h.model,
		nil,
		texts,
	)
	if err != nil {
		return nil, err
	}

	embeddings, err := h.featureExtraction(ctx, texts)
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
