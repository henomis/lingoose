package observer

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

type LLMObserver interface {
	Generation(*observer.Generation) (*observer.Generation, error)
	GenerationEnd(*observer.Generation) (*observer.Generation, error)
}

func StartObserveGeneration(
	ctx context.Context,
	name string,
	modelName string,
	ModelParameters types.M,
	t *thread.Thread,
) (*observer.Generation, error) {
	o, ok := observer.ContextValueObserverInstance(ctx).(LLMObserver)
	if o == nil || !ok {
		// No observer instance in context
		//nolint:nilnil
		return nil, nil
	}

	generation, err := o.Generation(
		&observer.Generation{
			TraceID:         observer.ContextValueTraceID(ctx),
			ParentID:        observer.ContextValueParentID(ctx),
			Name:            fmt.Sprintf("llm-%s", name),
			Model:           modelName,
			ModelParameters: ModelParameters,
			Input:           t.Messages,
		},
	)
	if err != nil {
		return nil, err
	}
	return generation, nil
}

func StopObserveGeneration(
	ctx context.Context,
	generation *observer.Generation,
	t *thread.Thread,
) error {
	o, ok := observer.ContextValueObserverInstance(ctx).(LLMObserver)
	if o == nil || !ok {
		// No observer instance in context
		return nil
	}

	generation.Output = t.LastMessage()
	_, err := o.GenerationEnd(generation)
	return err
}
