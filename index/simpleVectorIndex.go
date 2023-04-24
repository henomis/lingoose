package index

import (
	"encoding/json"
	"math"
	"os"
	"sort"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
)

type SimpleVectorIndexData struct {
	Document  document.Document  `json:"document"`
	Embedding embedder.Embedding `json:"embedding"`
}

type SimpleVectorIndex struct {
	Data []SimpleVectorIndexData `json:"data"`
}

func NewSimpleVectorIndex(documents []document.Document, embeddings []embedder.Embedding) *SimpleVectorIndex {

	simpleVectorIndex := &SimpleVectorIndex{}

	for i, document := range documents {
		simpleVectorIndex.Data = append(simpleVectorIndex.Data, SimpleVectorIndexData{
			Document:  document,
			Embedding: embeddings[i],
		})
	}

	return simpleVectorIndex
}

func (e SimpleVectorIndex) Save(filename string) error {

	jsonContent, err := json.Marshal(e)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonContent, 0644)
}

func (e *SimpleVectorIndex) Load(filename string) error {

	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(content, &e)
}

func (s *SimpleVectorIndex) Search(embeddingVector embedder.Embedding, topK *int) []SearchResponse {

	scores := s.cosineSimilarityBatch(embeddingVector)

	searchResponses := make([]SearchResponse, len(scores))

	for i, score := range scores {
		searchResponses[i] = SearchResponse{
			Document: s.Data[i].Document,
			Score:    score,
			Index:    i,
		}
	}

	//sort by similarity score
	sort.Slice(searchResponses, func(i, j int) bool {
		return searchResponses[i].Score > searchResponses[j].Score
	})

	//return topK
	if topK == nil {
		return searchResponses
	}

	return searchResponses[:*topK]

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
