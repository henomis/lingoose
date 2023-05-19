// Package pipeline provides a way to chain multiple llm executions.
package pipeline

import (
	"context"
	"errors"

	"github.com/henomis/lingoose/types"
)

var (
	ErrDecoding       = errors.New("decoding input error")
	ErrInvalidLmmMode = errors.New("invalid LLM mode")
	ErrLLMExecution   = errors.New("llm execution error")
)

const (
	NextTubeKey  = "next_tube"
	NextTubeExit = -1
)

type Memory interface {
	Get(key string) interface{}
	Set(key string, value interface{}) error
	All() types.M
	Delete(key string) error
	Clear() error
}

type Pipe interface {
	Run(ctx context.Context, input types.M) (types.M, error)
}

type PipelineCallback func(input types.M) (types.M, error)

type pipeline struct {
	pipes     []Pipe
	callbacks []PipelineCallback
}

func New(pipes ...Pipe) *pipeline {
	return &pipeline{
		pipes: pipes,
	}
}

func (p *pipeline) WithCallbacks(callbacks ...PipelineCallback) pipeline {
	p.callbacks = callbacks
	return *p
}

// Run chains the steps of the pipeline and returns the output of the last step.
func (p pipeline) Run(ctx context.Context, input types.M) (types.M, error) {
	var err error
	var output types.M
	currentTube := -1

	for {

		if currentTube == -1 {
			currentTube = 0
			output = input
		}

		output, err = p.pipes[currentTube].Run(ctx, output)
		if err != nil {
			return nil, err
		}

		if p.thereIsAValidCallbackForTube(currentTube) {
			output, err = p.callbacks[currentTube](output)
			if err != nil {
				return nil, err
			}

			nextTube := p.getNextTube(output)

			if nextTube != nil && *nextTube == NextTubeExit {
				break
			} else if nextTube != nil {
				currentTube = *nextTube
				continue
			}

		}

		currentTube++

		if currentTube == len(p.pipes) {
			break
		}

	}

	return output, nil
}

func SetNextTube(output types.M, nextTube int) types.M {
	output[NextTubeKey] = nextTube
	return output
}

func SetNextTubeExit(output types.M) types.M {
	output[NextTubeKey] = NextTubeExit
	return output
}

func (p *pipeline) thereIsAValidCallbackForTube(currentTube int) bool {
	return len(p.callbacks) == len(p.pipes) && p.callbacks[currentTube] != nil
}

func (p *pipeline) getNextTube(output types.M) *int {

	nextTube, ok := output[NextTubeKey]
	if !ok {
		return nil
	}

	currentTube, ok := nextTube.(int)
	if ok {
		return &currentTube
	}

	return nil

}
