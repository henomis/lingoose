package index

import (
	"math"

	"github.com/henomis/lingoose/embedding"
)

type Similarity struct {
	Score float32
	Index int
}

func cosineSimilarity(embeddingObjectA embedding.EmbeddingObject, embeddingObjectB embedding.EmbeddingObject) float32 {
	dotProduct := float32(0.0)
	normA := float32(0.0)
	normB := float32(0.0)

	for i := 0; i < len(embeddingObjectA.Vector); i++ {
		dotProduct += embeddingObjectA.Vector[i] * embeddingObjectB.Vector[i]
		normA += embeddingObjectA.Vector[i] * embeddingObjectA.Vector[i]
		normB += embeddingObjectB.Vector[i] * embeddingObjectB.Vector[i]
	}

	if normA == 0 || normB == 0 {
		return float32(0.0)
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

func cosineSimilarityBatch(embedding embedding.EmbeddingObject, embeddings []embedding.EmbeddingObject) []Similarity {

	similarities := make([]Similarity, len(embeddings))

	for i, e := range embeddings {
		similarities[i].Index = e.Index
		similarities[i].Score = cosineSimilarity(embedding, e)
	}

	return similarities
}
