// Package pipeline provides a way to chain multiple llm executions.
package pipeline

import (
	"context"
	"errors"

	"github.com/henomis/lingoose/types"
)

var (
	ErrDecoding       = errors.New("decoding input error")
	ErrInvalidLLMMode = errors.New("invalid LLM mode")
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

type History interface {
	Add(content string, meta types.Meta) error
}

type Pipe interface {
	Run(ctx context.Context, input types.M) (types.M, error)
}

type PipelineCallback func(ctx context.Context, values types.M) (types.M, error)

type Pipeline struct {
	pipes         map[int]Pipe
	preCallbacks  map[int]PipelineCallback
	postCallbacks map[int]PipelineCallback
}

func New(pipes ...Pipe) *Pipeline {

	pipesMap := make(map[int]Pipe)
	for i, pipe := range pipes {
		pipesMap[i] = pipe
	}

	return &Pipeline{
		pipes: pipesMap,
	}
}

func (p *Pipeline) WithPreCallbacks(callbacks ...PipelineCallback) *Pipeline {

	p.preCallbacks = make(map[int]PipelineCallback)
	for i, callback := range callbacks {
		p.preCallbacks[i] = callback
	}

	return p
}

func (p *Pipeline) WithPostCallbacks(callbacks ...PipelineCallback) *Pipeline {

	p.postCallbacks = make(map[int]PipelineCallback)
	for i, callback := range callbacks {
		p.postCallbacks[i] = callback
	}

	return p
}

// Run chains the steps of the pipeline and returns the output of the last step.
func (p Pipeline) Run(ctx context.Context, input types.M) (types.M, error) {
	var err error
	currentTube := 0

	if input == nil {
		input = types.M{}
	}

	output := input

	for {

		if p.thereIsAValidPreCallbackForTube(currentTube) {
			output, err = p.preCallbacks[currentTube](ctx, output)
			if err != nil {
				return nil, err
			}
		}

		output, err = p.pipes[currentTube].Run(ctx, output)
		if err != nil {
			return nil, err
		}

		if p.thereIsAValidPostCallbackForTube(currentTube) {
			output, err = p.postCallbacks[currentTube](ctx, output)
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

func (p *Pipeline) thereIsAValidPreCallbackForTube(currentTube int) bool {
	cb, ok := p.preCallbacks[currentTube]
	return cb != nil && ok
}

func (p *Pipeline) thereIsAValidPostCallbackForTube(currentTube int) bool {
	cb, ok := p.postCallbacks[currentTube]
	return cb != nil && ok
}

func (p *Pipeline) getNextTube(output types.M) *int {

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
