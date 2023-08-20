package index

import (
	"context"
	"sort"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
	"github.com/henomis/lingoose/types"
)

var (
	ErrInternal = "internal index error"
)

const (
	DefaultKeyID      = "id"
	DefaultKeyContent = "content"
)

type SearchResponse struct {
	ID       string
	Document document.Document
	Score    float64
}

type SearchResponses []SearchResponse

func (s SearchResponses) ToDocuments() []document.Document {
	documents := make([]document.Document, len(s))
	for i, searchResponse := range s {
		documents[i] = searchResponse.Document
	}
	return documents
}

type Embedder interface {
	Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error)
}

func FilterSearchResponses(searchResponses SearchResponses, topK int) SearchResponses {
	//sort by similarity score
	sort.Slice(searchResponses, func(i, j int) bool {
		return searchResponses[i].Score > searchResponses[j].Score
	})

	maxTopK := topK
	if maxTopK > len(searchResponses) {
		maxTopK = len(searchResponses)
	}

	return searchResponses[:maxTopK]
}

func DeepCopyMetadata(metadata types.Meta) types.Meta {
	metadataCopy := make(types.Meta)
	for k, v := range metadata {
		metadataCopy[k] = v
	}
	return metadataCopy
}
