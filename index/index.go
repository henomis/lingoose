package index

import (
	"context"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
)

type SearchResponse struct {
	Document document.Document
	Score    float32
	Index    int
}

type Embedder interface {
	Embed(ctx context.Context, docs []document.Document) ([]embedder.Embedding, error)
}
