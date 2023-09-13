package simplevectorindex

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/option"
	"github.com/henomis/lingoose/types"
)

const (
	defaultBatchSize = 32
	defaultTopK      = 10
)

type data struct {
	ID       string     `json:"id"`
	Metadata types.Meta `json:"metadata"`
	Values   []float64  `json:"values"`
}

type Index struct {
	data       []data
	outputPath string
	name       string
	embedder   index.Embedder
}

type SimpleVectorIndexFilterFn func([]index.SearchResult) []index.SearchResult

func New(name string, outputPath string, embedder index.Embedder) *Index {
	simpleVectorIndex := &Index{
		data:       []data{},
		outputPath: outputPath,
		name:       name,
		embedder:   embedder,
	}

	return simpleVectorIndex
}

func (s *Index) LoadFromDocuments(ctx context.Context, documents []document.Document) error {
	err := s.load()
	if err != nil {
		return fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	id := 0
	for i := 0; i < len(documents); i += defaultBatchSize {

		end := i + defaultBatchSize
		if end > len(documents) {
			end = len(documents)
		}

		texts := []string{}
		for _, document := range documents[i:end] {
			texts = append(texts, document.Content)
		}

		embeddings, err := s.embedder.Embed(ctx, texts)
		if err != nil {
			return fmt.Errorf("%s: %w", index.ErrInternal, err)
		}

		for j, document := range documents[i:end] {
			s.data = append(s.data, buildDataFromEmbeddingAndDocument(id, embeddings[j], document))
			id++
		}

	}

	err = s.save()
	if err != nil {
		return fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	return nil
}

func buildDataFromEmbeddingAndDocument(
	id int,
	embedding embedder.Embedding,
	document document.Document,
) data {
	metadata := index.DeepCopyMetadata(document.Metadata)
	metadata[index.DefaultKeyContent] = document.Content
	return data{
		ID:       fmt.Sprintf("%d", id),
		Values:   embedding,
		Metadata: metadata,
	}
}

func (s Index) save() error {

	jsonContent, err := json.Marshal(s.data)
	if err != nil {
		return err
	}

	return os.WriteFile(s.database(), jsonContent, 0644)
}

func (s *Index) load() error {
	if len(s.data) > 0 {
		return nil
	}

	if _, err := os.Stat(s.database()); os.IsNotExist(err) {
		return s.save()
	}

	content, err := os.ReadFile(s.database())
	if err != nil {
		return err
	}

	return json.Unmarshal(content, &s.data)
}

func (s *Index) database() string {
	return strings.Join([]string{s.outputPath, s.name + ".json"}, string(os.PathSeparator))
}

func (s *Index) IsEmpty() (bool, error) {

	err := s.load()
	if err != nil {
		return true, fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	return len(s.data) == 0, nil
}

func (s *Index) Add(ctx context.Context, item *index.Data) error {
	err := s.load()
	if err != nil {
		return fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	if item.ID == "" {
		lastID := s.data[len(s.data)-1].ID
		lastIDAsInt, err := strconv.Atoi(lastID)
		if err != nil {
			return fmt.Errorf("%s: %w", index.ErrInternal, err)
		}

		item.ID = fmt.Sprintf("%d", lastIDAsInt+1)
	}

	s.data = append(
		s.data,
		data{
			ID:       item.ID,
			Values:   item.Values,
			Metadata: item.Metadata,
		},
	)

	return s.save()
}

func (s *Index) Search(ctx context.Context, values []float64, opts ...option.Option) (index.SearchResults, error) {
	sviOptions := &option.Options{
		TopK: defaultTopK,
	}

	for _, opt := range opts {
		opt(sviOptions)
	}

	err := s.load()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	return s.similaritySearch(ctx, values, sviOptions)
}

func (s *Index) Query(ctx context.Context, query string, opts ...option.Option) (index.SearchResults, error) {

	sviOptions := &option.Options{
		TopK: defaultTopK,
	}

	for _, opt := range opts {
		opt(sviOptions)
	}

	err := s.load()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	embeddings, err := s.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	return s.similaritySearch(ctx, embeddings[0], sviOptions)
}

func (s *Index) similaritySearch(
	ctx context.Context,
	embedding embedder.Embedding,
	opts *option.Options,
) (index.SearchResults, error) {

	scores := s.cosineSimilarityBatch(embedding)

	searchResults := make([]index.SearchResult, len(scores))

	for i, score := range scores {
		searchResults[i] = index.SearchResult{
			Data: index.Data{
				ID:       s.data[i].ID,
				Values:   s.data[i].Values,
				Metadata: s.data[i].Metadata,
			},
			Score: score,
		}
	}

	if opts.Filter != nil {
		searchResults = opts.Filter.(SimpleVectorIndexFilterFn)(searchResults)
	}

	return index.FilterSearchResults(searchResults, opts.TopK), nil
}

func (s *Index) cosineSimilarity(a embedder.Embedding, b embedder.Embedding) float64 {
	dotProduct := float64(0.0)
	normA := float64(0.0)
	normB := float64(0.0)

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return float64(0.0)
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (s *Index) cosineSimilarityBatch(a embedder.Embedding) []float64 {
	scores := make([]float64, len(s.data))

	for i := range s.data {
		scores[i] = s.cosineSimilarity(a, s.data[i].Values)
	}

	return scores
}
