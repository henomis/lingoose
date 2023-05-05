package openaiembedder

import (
	"math"

	"github.com/henomis/lingoose/embedder"
)

func normalizeEmbeddings(embeddings []embedder.Embedding, lens []float64) embedder.Embedding {

	chunkAvgEmbeddings := average(embeddings, lens)
	norm := norm(chunkAvgEmbeddings)

	for i := range chunkAvgEmbeddings {
		chunkAvgEmbeddings[i] = chunkAvgEmbeddings[i] / norm
	}

	return chunkAvgEmbeddings
}

func average(embeddings []embedder.Embedding, lens []float64) embedder.Embedding {
	average := make(embedder.Embedding, len(embeddings[0]))
	totalWeight := 0.0

	for i, embedding := range embeddings {
		weight := lens[i]
		totalWeight += weight
		for j, v := range embedding {
			average[j] += v * weight
		}
	}

	for i, v := range average {
		average[i] = v / totalWeight
	}

	return average
}

func norm(a embedder.Embedding) float64 {
	var sum float64
	for _, v := range a {
		sum += math.Pow(v, 2)
	}
	return math.Sqrt(sum)
}
