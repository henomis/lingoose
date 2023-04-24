package index

import (
	"encoding/json"
	"os"

	"github.com/henomis/lingoose/embedding"
)

type VectorIndex struct {
	embeddingObjects []embedding.EmbeddingObject
}

func NewVectorIndex(objects []embedding.EmbeddingObject) *VectorIndex {
	return &VectorIndex{
		embeddingObjects: objects,
	}
}

func (e VectorIndex) Save(filename string) error {

	jsonContent, err := json.Marshal(e.embeddingObjects)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonContent, 0644)
}

func (e *VectorIndex) Load(filename string) error {

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
