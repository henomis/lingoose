package index

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
)

const (
	defaultBatchSize = 32
)

type simpleVectorIndexData struct {
	Document  document.Document  `json:"document"`
	Embedding embedder.Embedding `json:"embedding"`
}

type SimpleVectorIndex struct {
	data       []simpleVectorIndexData
	outputPath string
	name       string
	embedder   Embedder
}

func NewSimpleVectorIndex(name string, outputPath string, embedder Embedder) *SimpleVectorIndex {
	simpleVectorIndex := &SimpleVectorIndex{
		data:       []simpleVectorIndexData{},
		outputPath: outputPath,
		name:       name,
		embedder:   embedder,
	}

	return simpleVectorIndex
}

func (s *SimpleVectorIndex) LoadFromDocuments(ctx context.Context, documents []document.Document) error {

	s.data = []simpleVectorIndexData{}

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
			return fmt.Errorf("%s: %w", ErrInternal, err)
		}

		for j, document := range documents[i:end] {
			s.data = append(s.data, simpleVectorIndexData{
				Document:  document,
				Embedding: embeddings[j],
			})

			documents[documentIndex].Metadata[defaultKeyID] = fmt.Sprintf("%d", documentIndex)
			documentIndex++
		}

	}

	err := s.save()
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInternal, err)
	}

	return nil
}

func (s SimpleVectorIndex) save() error {

	jsonContent, err := json.Marshal(s.data)
	if err != nil {
		return err
	}

	return os.WriteFile(s.database(), jsonContent, 0644)
}

func (s *SimpleVectorIndex) load() error {

	content, err := os.ReadFile(s.database())
	if err != nil {
		return err
	}

	return json.Unmarshal(content, &s.data)
}

func (s *SimpleVectorIndex) database() string {
	return strings.Join([]string{s.outputPath, s.name + ".json"}, string(os.PathSeparator))
}

func (s *SimpleVectorIndex) IsEmpty() (bool, error) {

	err := s.load()
	if err != nil {
		return true, fmt.Errorf("%s: %w", ErrInternal, err)
	}

	return len(s.data) == 0, nil
}

func (s *SimpleVectorIndex) SimilaritySearch(ctx context.Context, query string, topK *int) (SearchResponses, error) {

	err := s.load()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrInternal, err)
	}

	embeddings, err := s.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrInternal, err)
	}

	scores := s.cosineSimilarityBatch(embeddings[0])

	searchResponses := make([]SearchResponse, len(scores))

	for i, score := range scores {

		id := s.data[i].Document.Metadata[defaultKeyID].(string)

		searchResponses[i] = SearchResponse{
			ID:       id,
			Document: s.data[i].Document,
			Score:    score,
		}
	}

	return filterSearchResponses(searchResponses, topK), nil
}

func (s *SimpleVectorIndex) cosineSimilarity(a embedder.Embedding, b embedder.Embedding) float64 {
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

func (s *SimpleVectorIndex) cosineSimilarityBatch(a embedder.Embedding) []float64 {

	scores := make([]float64, len(s.data))

	for i := range s.data {
		scores[i] = s.cosineSimilarity(a, s.data[i].Embedding)
	}

	return scores
}
