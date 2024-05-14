package observer

import (
	"fmt"

	obs "github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

type LLMObserver interface {
	Generation(*obs.Generation) (*obs.Generation, error)
	GenerationEnd(*obs.Generation) (*obs.Generation, error)
}

func StartObserveGeneration(
	o LLMObserver,
	name string,
	modelName string,
	ModelParameters types.M,
	traceID string,
	parentID string,
	t *thread.Thread,
) (*obs.Generation, error) {
	generation, err := o.Generation(
		&obs.Generation{
			TraceID:         traceID,
			ParentID:        parentID,
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
	o LLMObserver,
	generation *obs.Generation,
	t *thread.Thread,
) error {
	generation.Output = t.LastMessage()
	_, err := o.GenerationEnd(generation)
	return err
}
