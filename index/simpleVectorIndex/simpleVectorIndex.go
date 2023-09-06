package simplevectorindex

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/option"
)

const (
	defaultBatchSize = 32
	defaultTopK      = 10
)

type data struct {
	Document  document.Document  `json:"document"`
	Embedding embedder.Embedding `json:"embedding"`
}

type Index struct {
	data       []data
	outputPath string
	name       string
	embedder   index.Embedder
}

type SimpleVectorIndexFilterFn func([]index.SearchResponse) []index.SearchResponse

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

	s.data = []data{}

	documentIndex := 0
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
			s.data = append(s.data, data{
				Document:  document,
				Embedding: embeddings[j],
			})

			documents[documentIndex].Metadata[index.DefaultKeyID] = fmt.Sprintf("%d", documentIndex)
			documentIndex++
		}

	}

	err := s.save()
	if err != nil {
		return fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	return nil
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

func (s *Index) SimilaritySearch(ctx context.Context, query string, opts ...option.Option) (index.SearchResponses, error) {

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

	scores := s.cosineSimilarityBatch(embeddings[0])

	searchResponses := make([]index.SearchResponse, len(scores))

	for i, score := range scores {

		id := s.data[i].Document.Metadata[index.DefaultKeyID].(string)

		searchResponses[i] = index.SearchResponse{
			ID:       id,
			Document: s.data[i].Document,
			Score:    score,
		}
	}

	if sviOptions.Filter != nil {
		searchResponses = sviOptions.Filter.(SimpleVectorIndexFilterFn)(searchResponses)
	}

	return index.FilterSearchResponses(searchResponses, sviOptions.TopK), nil
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
		scores[i] = s.cosineSimilarity(a, s.data[i].Embedding)
	}

	return scores
}
