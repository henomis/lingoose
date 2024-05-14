package observer

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/embedder"
	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/types"
)

type EmbeddingObserver interface {
	Embedding(*observer.Embedding) (*observer.Embedding, error)
	EmbeddingEnd(*observer.Embedding) (*observer.Embedding, error)
}

func StartObserveEmbedding(
	ctx context.Context,
	name string,
	modelName string,
	ModelParameters types.M,
	texts []string,
) (*observer.Embedding, error) {
	o, ok := observer.ContextValueObserverInstance(ctx).(EmbeddingObserver)
	if o == nil || !ok {
		// No observer instance in context
		//nolint:nilnil
		return nil, nil
	}

	embedding, err := o.Embedding(
		&observer.Embedding{
			TraceID:         observer.ContextValueTraceID(ctx),
			ParentID:        observer.ContextValueParentID(ctx),
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
	ctx context.Context,
	embedding *observer.Embedding,
	embeddings []embedder.Embedding,
) error {
	o, ok := observer.ContextValueObserverInstance(ctx).(EmbeddingObserver)
	if o == nil || !ok {
		// No observer instance in context
		return nil
	}

	embedding.Output = embeddings
	_, err := o.EmbeddingEnd(embedding)
	return err
}
