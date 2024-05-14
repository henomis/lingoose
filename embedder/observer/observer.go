package observer

import (
	"fmt"

	"github.com/henomis/lingoose/embedder"
	obs "github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/types"
)

type EmbeddingObserver interface {
	Embedding(*obs.Embedding) (*obs.Embedding, error)
	EmbeddingEnd(*obs.Embedding) (*obs.Embedding, error)
}

func StartObserveEmbedding(
	o EmbeddingObserver,
	name string,
	modelName string,
	ModelParameters types.M,
	traceID string,
	parentID string,
	texts []string,
) (*obs.Embedding, error) {
	embedding, err := o.Embedding(
		&obs.Embedding{
			TraceID:         traceID,
			ParentID:        parentID,
			Name:            fmt.Sprintf("embedding-%s", name),
			Model:           modelName,
			ModelParameters: ModelParameters,
			Input:           texts,
		},
	)
	if err != nil {
		return nil, err
	}
	return embedding, nil
}

func StopObserveEmbedding(
	o EmbeddingObserver,
	embedding *obs.Embedding,
	embeddings []embedder.Embedding,
) error {
	embedding.Output = embeddings
	_, err := o.EmbeddingEnd(embedding)
	return err
}
