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

type SimpleVectorIndexData struct {
	Document  document.Document  `json:"document"`
	Embedding embedder.Embedding `json:"embedding"`
}

type SimpleVectorIndex struct {
	Data       []SimpleVectorIndexData `json:"data"`
	outputPath string                  `json:"-"`
	name       string                  `json:"-"`
	embedder   Embedder                `json:"-"`
}

func NewSimpleVectorIndex(name string, outputPath string, embedder Embedder) (*SimpleVectorIndex, error) {
	simpleVectorIndex := &SimpleVectorIndex{
		Data:       []SimpleVectorIndexData{},
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

func (s *SimpleVectorIndex) LoadFromDocuments(ctx context.Context, documents []document.Document) error {

	embeddings, err := s.embedder.Embed(ctx, documents)
	if err != nil {
		return err
	}

	s.Data = []SimpleVectorIndexData{}

	for i, document := range documents {
		s.Data = append(s.Data, SimpleVectorIndexData{
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

func (s SimpleVectorIndex) save() error {

	jsonContent, err := json.Marshal(s)
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

	return json.Unmarshal(content, &s)
}

func (s *SimpleVectorIndex) database() string {
	return strings.Join([]string{s.outputPath, s.name + ".json"}, string(os.PathSeparator))
}

func (s *SimpleVectorIndex) Size() (int64, error) {
	return int64(len(s.Data)), nil
}

func (s *SimpleVectorIndex) SimilaritySearch(ctx context.Context, query string, topK *int) ([]SearchResponse, error) {

	embeddings, err := s.embedder.Embed(ctx, []document.Document{{Content: query}})
	if err != nil {
		return nil, err
	}

	scores := s.cosineSimilarityBatch(embeddings[0])

	searchResponses := make([]SearchResponse, len(scores))

	for i, score := range scores {

		id := s.Data[i].Document.Metadata[defaultKeyID].(string)

		searchResponses[i] = SearchResponse{
			ID:       id,
			Document: s.Data[i].Document,
			Score:    score,
		}
	}

	return filterSearchResponses(searchResponses, topK), nil
}

func (s *SimpleVectorIndex) cosineSimilarity(a embedder.Embedding, b embedder.Embedding) float32 {
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

func (s *SimpleVectorIndex) cosineSimilarityBatch(a embedder.Embedding) []float32 {

	scores := make([]float32, len(s.Data))

	for i := range s.Data {
		scores[i] = s.cosineSimilarity(a, s.Data[i].Embedding)
	}

	return scores
}
