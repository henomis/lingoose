package observer

import (
	"fmt"

	obs "github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

type LLMObserver interface {
	Span(*obs.Span) (*obs.Span, error)
	SpanEnd(*obs.Span) (*obs.Span, error)
	Generation(*obs.Generation) (*obs.Generation, error)
	GenerationEnd(*obs.Generation) (*obs.Generation, error)
}

func StartObserveGeneration(
	o LLMObserver,
	name string,
	modelName string,
	ModelParameters types.M,
	traceID string,
	t *thread.Thread,
) (*obs.Span, *obs.Generation, error) {
	span, err := o.Span(
		&obs.Span{
			TraceID: traceID,
			Name:    name,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	generation, err := o.Generation(
		&obs.Generation{
			TraceID:         traceID,
			ParentID:        span.ID,
			Name:            fmt.Sprintf("%s-%s", name, modelName),
			Model:           modelName,
			ModelParameters: ModelParameters,
			Input:           t.Messages,
		},
	)
	if err != nil {
		return nil, nil, err
	}
	return span, generation, nil
}

func StopObserveGeneration(
	o LLMObserver,
	span *obs.Span,
	generation *obs.Generation,
	t *thread.Thread,
) error {
	_, err := o.SpanEnd(span)
	if err != nil {
		return err
	}

	generation.Output = t.LastMessage()
	_, err = o.GenerationEnd(generation)
	return err
}
