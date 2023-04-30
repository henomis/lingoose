package index

import (
	"context"
	"sort"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
)

var (
	ErrInternal = "internal index error"
)

const (
	defaultKeyID      = "id"
	defaultKeyContent = "content"
)

type SearchResponse struct {
	ID       string
	Document document.Document
	Score    float32
}

type Embedder interface {
	Embed(ctx context.Context, docs []document.Document) ([]embedder.Embedding, error)
}

func filterSearchResponses(searchResponses []SearchResponse, topK *int) []SearchResponse {
	//sort by similarity score
	sort.Slice(searchResponses, func(i, j int) bool {
		return searchResponses[i].Score > searchResponses[j].Score
	})

	//return topK
	if topK == nil {
		return searchResponses
	}

	maxTopK := *topK
	if maxTopK > len(searchResponses) {
		maxTopK = len(searchResponses)
	}

	return searchResponses[:maxTopK]
}
