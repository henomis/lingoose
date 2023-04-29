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

type simpleVectorIndexData struct {
	Document  document.Document  `json:"document"`
	Embedding embedder.Embedding `json:"embedding"`
}

type simpleVectorIndex struct {
	data       []simpleVectorIndexData
	outputPath string
	name       string
	embedder   Embedder
}

func NewSimpleVectorIndex(name string, outputPath string, embedder Embedder) (*simpleVectorIndex, error) {
	simpleVectorIndex := &simpleVectorIndex{
		data:       []simpleVectorIndexData{},
		outputPath: outputPath,
		name:       name,
		embedder:   embedder,
	}

	_, err := os.Stat(simpleVectorIndex.database())
	if err == nil {
		err = simpleVectorIndex.load()
		if err != nil {
			return nil, err
		}
	}

	return simpleVectorIndex, nil
}

func (s *simpleVectorIndex) LoadFromDocuments(ctx context.Context, documents []document.Document) error {

	embeddings, err := s.embedder.Embed(ctx, documents)
	if err != nil {
		return err
	}

	s.data = []simpleVectorIndexData{}

	for i, document := range documents {
		s.data = append(s.data, simpleVectorIndexData{
			Document:  document,
			Embedding: embeddings[i],
		})

		documents[i].Metadata[defaultKeyID] = fmt.Sprintf("%d", i)
	}

	err = s.save()
	if err != nil {
		return err
	}

	return nil
}

func (s simpleVectorIndex) save() error {

	jsonContent, err := json.Marshal(s.data)
	if err != nil {
		return err
	}

	return os.WriteFile(s.database(), jsonContent, 0644)
}

func (s *simpleVectorIndex) load() error {

	content, err := os.ReadFile(s.database())
	if err != nil {
		return err
	}

	return json.Unmarshal(content, &s.data)
}

func (s *simpleVectorIndex) database() string {
	return strings.Join([]string{s.outputPath, s.name + ".json"}, string(os.PathSeparator))
}

func (s *simpleVectorIndex) IsEmpty() (bool, error) {
	return len(s.data) == 0, nil
}

func (s *simpleVectorIndex) SimilaritySearch(ctx context.Context, query string, topK *int) ([]SearchResponse, error) {

	err := s.load()
	if err != nil {
		return nil, err
	}

	embeddings, err := s.embedder.Embed(ctx, []document.Document{{Content: query}})
	if err != nil {
		return nil, err
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

func (s *simpleVectorIndex) cosineSimilarity(a embedder.Embedding, b embedder.Embedding) float32 {
	dotProduct := float32(0.0)
	normA := float32(0.0)
	normB := float32(0.0)

	for i := 0; i < len(a.Embedding); i++ {
		dotProduct += a.Embedding[i] * b.Embedding[i]
		normA += a.Embedding[i] * a.Embedding[i]
		normB += b.Embedding[i] * b.Embedding[i]
	}

	if normA == 0 || normB == 0 {
		return float32(0.0)
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

func (s *simpleVectorIndex) cosineSimilarityBatch(a embedder.Embedding) []float32 {

	scores := make([]float32, len(s.data))

	for i := range s.data {
		scores[i] = s.cosineSimilarity(a, s.data[i].Embedding)
	}

	return scores
}
