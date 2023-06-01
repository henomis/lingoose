package huggingfaceembedder

import (
	"context"
	"os"

	"github.com/henomis/lingoose/embedder"
)

const (
	hfDefaultEmbedderModel = "sentence-transformers/all-MiniLM-L6-v2"
)

type huggingFaceEmbedder struct {
	token string
	model string
}

func New() *huggingFaceEmbedder {
	return &huggingFaceEmbedder{
		token: os.Getenv("HUGGING_FACE_HUB_TOKEN"),
		model: hfDefaultEmbedderModel,
	}
}

func (h *huggingFaceEmbedder) WithToken(token string) *huggingFaceEmbedder {
	h.token = token
	return h
}

func (h *huggingFaceEmbedder) WithModel(model string) *huggingFaceEmbedder {
	h.model = model
	return h
}

func (h *huggingFaceEmbedder) Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	return h.featureExtraction(ctx, texts)
}
