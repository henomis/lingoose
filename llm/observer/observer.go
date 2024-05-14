package observer

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/observer"
	obs "github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

type LLMObserver interface {
	Generation(*obs.Generation) (*obs.Generation, error)
	GenerationEnd(*obs.Generation) (*obs.Generation, error)
}

func StartObserveGeneration(
	// o LLMObserver,
	ctx context.Context,
	name string,
	modelName string,
	ModelParameters types.M,
	t *thread.Thread,
) (*obs.Generation, error) {
	o, ok := observer.ContextValueObserverInstance(ctx).(LLMObserver)
	if o == nil || !ok {
		// No observer instance in context
		return nil, nil
	}

	generation, err := o.Generation(
		&obs.Generation{
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
	generation *obs.Generation,
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
