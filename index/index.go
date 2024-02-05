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
	defaultIncludeContent  = true
)

type AddDataCallback func(data *Data) error

type Data struct {
	ID       string
	Values   []float64
	Metadata types.Meta
}

type Embedder interface {
	Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error)
}

type VectorDB interface {
	Insert(context.Context, []Data) error
	IsEmpty(context.Context) (bool, error)
	Search(context.Context, []float64, *option.Options) (SearchResults, error)
	Drop(ctx context.Context) error
	Delete(ctx context.Context, ids []string) error
}

type Index struct {
	vectorDB        VectorDB
	embedder        Embedder
	batchInsertSize int
	includeContent  bool
	addDataCallback AddDataCallback
}

func New(vectorDB VectorDB, embedder Embedder) *Index {
	return &Index{
		vectorDB:        vectorDB,
		embedder:        embedder,
		batchInsertSize: defaultBatchInsertSize,
		includeContent:  defaultIncludeContent,
		addDataCallback: nil,
	}
}

func (i *Index) WithIncludeContents(includeContents bool) *Index {
	i.includeContent = includeContents
	return i
}

func (i *Index) WithBatchInsertSize(batchInsertSize int) *Index {
	i.batchInsertSize = batchInsertSize
	return i
}

// WithAddDataCallback allows to modify the data before it is added to the index.
// This can be useful to add additional metadata to the vector.
func (i *Index) WithAddDataCallback(callback AddDataCallback) *Index {
	i.addDataCallback = callback
	return i
}

func (i *Index) LoadFromDocuments(ctx context.Context, documents []document.Document) error {
	err := i.batchUpsert(ctx, documents)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInternal, err)
	}
	return nil
}

func (i *Index) Add(ctx context.Context, data *Data) error {
	if data == nil {
		return nil
	}

	if i.addDataCallback != nil {
		err := i.addDataCallback(data)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInternal, err)
		}
	}

	return i.vectorDB.Insert(ctx, []Data{*data})
}

func (i *Index) IsEmpty(ctx context.Context) (bool, error) {
	return i.vectorDB.IsEmpty(ctx)
}

func (i *Index) Drop(ctx context.Context) error {
	return i.vectorDB.Drop(ctx)
}

func (i *Index) Search(ctx context.Context, values []float64, opts ...option.Option) (SearchResults, error) {
	options := &option.Options{
		TopK: defaultTopK,
	}

	for _, opt := range opts {
		opt(options)
	}
	return i.vectorDB.Search(ctx, values, options)
}

func (i *Index) Query(ctx context.Context, query string, opts ...option.Option) (SearchResults, error) {
	embeddings, err := i.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	return i.Search(ctx, embeddings[0], opts...)
}

func (i *Index) Embedder() Embedder {
	return i.embedder
}

func (i *Index) batchUpsert(ctx context.Context, documents []document.Document) error {
	for j := 0; j < len(documents); j += i.batchInsertSize {
		batchEnd := j + i.batchInsertSize
		if batchEnd > len(documents) {
			batchEnd = len(documents)
		}

		texts := []string{}
		for _, document := range documents[j:batchEnd] {
			texts = append(texts, document.Content)
		}

		embeddings, err := i.embedder.Embed(ctx, texts)
		if err != nil {
			return err
		}

		data, err := i.buildDataFromEmbeddingsAndDocuments(embeddings, documents, j)
		if err != nil {
			return err
		}

		if i.addDataCallback != nil {
			for j := range data {
				callbackErr := i.addDataCallback(&data[j])
				if callbackErr != nil {
					return fmt.Errorf("%w: %w", ErrInternal, callbackErr)
				}
			}
		}

		err = i.vectorDB.Insert(ctx, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Index) buildDataFromEmbeddingsAndDocuments(
	embeddings []embedder.Embedding,
	documents []document.Document,
	startIndex int,
) ([]Data, error) {
	var vectors []Data

	for j, embedding := range embeddings {
		metadata := DeepCopyMetadata(documents[startIndex+j].Metadata)

		// inject document content into vector metadata
		if i.includeContent {
			metadata[DefaultKeyContent] = documents[startIndex+j].Content
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
		documents[startIndex+j].Metadata[DefaultKeyID] = vectorID.String()
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

func GetDefaultOptions() *option.Options {
	return &option.Options{
		TopK: defaultTopK,
	}
}
