package index

import (
	"encoding/json"
	"os"
	"sort"

	"github.com/henomis/lingoose/embedding"
)

type SimpleVector struct {
	embeddingObjects []embedding.EmbeddingObject
}

func NewVectorIndex(objects []embedding.EmbeddingObject) *SimpleVector {
	return &SimpleVector{
		embeddingObjects: objects,
	}
}

func (e *SimpleVector) Embeddings() []embedding.EmbeddingObject {
	return e.embeddingObjects
}

func (e SimpleVector) Save(filename string) error {

	jsonContent, err := json.Marshal(e.embeddingObjects)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonContent, 0644)
}

func (e *SimpleVector) Load(filename string) error {

	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var objs []embedding.EmbeddingObject
	err = json.Unmarshal(content, &objs)
	if err != nil {
		return err
	}

	e.embeddingObjects = objs

	return nil
}

func (s *SimpleVector) Search(embeddingVector embedding.EmbeddingObject, topK *int) []Similarity {

	similarities := cosineSimilarityBatch(embeddingVector, s.embeddingObjects)

	//sort by similarity score

	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Score > similarities[j].Score
	})

	//return topK
	if topK == nil {
		return similarities
	}

	return similarities[:*topK]

}
