package index

import (
	"context"
	"errors"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
	"github.com/henomis/lingoose/types"
)

var (
	ErrInternal = errors.New("internal index error")
)

const (
	DefaultKeyID      = "id"
	DefaultKeyContent = "content"
)

type Data struct {
	ID       string
	Values   []float64
	Metadata types.Meta
}

type SearchResult struct {
	Data
	Score float64
}

func (s *SearchResult) Content() string {
	return s.Metadata[DefaultKeyContent].(string)
}

type SearchResults []SearchResult

func (s SearchResults) ToDocuments() []document.Document {
	documents := make([]document.Document, len(s))
	for i, searchResult := range s {
		metadata := DeepCopyMetadata(searchResult.Metadata)
		content, ok := metadata[DefaultKeyContent].(string)
		if !ok {
			content = ""
		}
		delete(metadata, DefaultKeyContent)

		documents[i] = document.Document{
			Content:  content,
			Metadata: metadata,
		}
	}
	return documents
}

type Embedder interface {
	Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error)
}

func DeepCopyMetadata(metadata types.Meta) types.Meta {
	metadataCopy := make(types.Meta)
	for k, v := range metadata {
		metadataCopy[k] = v
	}
	return metadataCopy
}
