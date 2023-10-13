package index

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
	"github.com/henomis/lingoose/index/option"
	"github.com/henomis/lingoose/types"
)

var (
	ErrInternal = errors.New("internal index error")
)

const (
	DefaultKeyID           = "id"
	DefaultKeyContent      = "content"
	defaultBatchInsertSize = 32
	defaultTopK            = 10
)

type Data struct {
	ID       string
	Values   []float64
	Metadata types.Meta
}

type Embedder interface {
	Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error)
}

type IndexEngine interface {
	Insert(context.Context, []Data) error
	IsEmpty(context.Context) (bool, error)
	Search(context.Context, []float64, *option.Options) (SearchResults, error)
}

type Index struct {
	indexEngine     IndexEngine
	embedder        Embedder
	batchInsertSize int
	includeContent  bool
}

func New(indexEngine IndexEngine, embedder Embedder) *Index {
	return &Index{
		indexEngine:     indexEngine,
		embedder:        embedder,
		batchInsertSize: defaultBatchInsertSize,
	}
}

func (i *Index) WithIncludeContents(includeContents bool) *Index {
	i.includeContent = true
	return i
}

func (i *Index) WithBatchInsertSize(batchInsertSize int) *Index {
	i.batchInsertSize = batchInsertSize
	return i
}

func (i *Index) LoadFromDocuments(ctx context.Context, documents []document.Document) error {
	err := i.batchUpsert(ctx, documents)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInternal, err)
	}
	return nil
}

func (i *Index) IsEmpty(ctx context.Context) (bool, error) {
	return i.indexEngine.IsEmpty(ctx)
}

func (i *Index) Search(ctx context.Context, values []float64, opts ...option.Option) (SearchResults, error) {
	options := &option.Options{
		TopK: defaultTopK,
	}

	for _, opt := range opts {
		opt(options)
	}
	return i.indexEngine.Search(ctx, values, options)
}

func (i *Index) Query(ctx context.Context, query string, opts ...option.Option) (SearchResults, error) {
	embeddings, err := i.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	return i.Search(ctx, embeddings[0], opts...)
}

func (q *Index) batchUpsert(ctx context.Context, documents []document.Document) error {
	for i := 0; i < len(documents); i += q.batchInsertSize {
		batchEnd := i + q.batchInsertSize
		if batchEnd > len(documents) {
			batchEnd = len(documents)
		}

		texts := []string{}
		for _, document := range documents[i:batchEnd] {
			texts = append(texts, document.Content)
		}

		embeddings, err := q.embedder.Embed(ctx, texts)
		if err != nil {
			return err
		}

		data, err := q.buildDataFromEmbeddingsAndDocuments(embeddings, documents, i)
		if err != nil {
			return err
		}

		err = q.indexEngine.Insert(ctx, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (q *Index) buildDataFromEmbeddingsAndDocuments(
	embeddings []embedder.Embedding,
	documents []document.Document,
	startIndex int,
) ([]Data, error) {
	var vectors []Data

	for i, embedding := range embeddings {
		metadata := DeepCopyMetadata(documents[startIndex+i].Metadata)

		// inject document content into vector metadata
		if q.includeContent {
			metadata[DefaultKeyContent] = documents[startIndex+i].Content
		}

		vectorID, err := uuid.NewUUID()
		if err != nil {
			return nil, err
		}

		vectors = append(vectors, Data{
			ID:       vectorID.String(),
			Values:   embedding,
			Metadata: metadata,
		})

		// inject vector ID into document metadata
		documents[startIndex+i].Metadata[DefaultKeyID] = vectorID.String()
	}

	return vectors, nil
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

func DeepCopyMetadata(metadata types.Meta) types.Meta {
	metadataCopy := make(types.Meta)
	for k, v := range metadata {
		metadataCopy[k] = v
	}
	return metadataCopy
}
