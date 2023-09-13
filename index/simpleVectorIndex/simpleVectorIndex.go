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
	err := s.load()
	if err != nil {
		return fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

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
			s.data = append(s.data, buildDataFromEmbeddingAndDocument(documentIndex, embeddings[j], document))
			documentIndex++
		}

	}

	err = s.save()
	if err != nil {
		return fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	return nil
}

func buildDataFromEmbeddingAndDocument(
	i int,
	embedding embedder.Embedding,
	document document.Document,
) data {
	metadata := index.DeepCopyMetadata(document.Metadata)
	metadata[index.DefaultKeyContent] = document.Content
	return data{
		ID:       fmt.Sprintf("%d", i),
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
		searchResponses[i] = index.SearchResponse{
			ID: s.data[i].ID,
			Document: document.Document{
				Content:  s.data[i].Metadata[index.DefaultKeyContent].(string),
				Metadata: s.data[i].Metadata,
			},
			Score: score,
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
		scores[i] = s.cosineSimilarity(a, s.data[i].Values)
	}

	return scores
}
