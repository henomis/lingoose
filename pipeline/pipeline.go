// Package pipeline provides a way to chain multiple llm executions.
package pipeline

import (
	"context"
	"errors"
	"fmt"

	"github.com/henomis/lingoose/types"
	"github.com/mitchellh/mapstructure"
)

var (
	ErrDecoding       = errors.New("decoding input error")
	ErrInvalidLmmMode = errors.New("invalid LLM mode")
	ErrLLMExecution   = errors.New("llm execution error")
)

type Memory interface {
	Get(key string) interface{}
	Set(key string, value interface{}) error
	All() types.M
	Delete(key string) error
	Clear() error
}

type Tube struct {
	llm       Llm
	decoder   Decoder
	namespace string
	memory    Memory
}

type Pipe interface {
	Run(ctx context.Context, input types.M) (types.M, error)
}

type pipeline struct {
	pipes []Pipe
}

func New(pipes ...Pipe) pipeline {
	return pipeline{
		pipes: pipes,
	}
}

func NewTube(
	llm Llm,
) *Tube {
	return &Tube{
		llm:     llm,
		decoder: &defaultDecoder{},
	}
}

func (t *Tube) Namespace() string {
	return t.namespace
}

func (t *Tube) WithMemory(namespace string, memory Memory) *Tube {
	t.namespace = namespace
	t.memory = memory
	return t
}

func (t *Tube) WithDecoder(decoder Decoder) *Tube {
	t.decoder = decoder
	return t
}

// Run execute the step and return the output.
// The prompt is formatted with the input and the output of the prompt is used as input for the LLM.
// If the step has a memory, the output is stored in the memory.
func (s *Tube) Run(ctx context.Context, input types.M) (types.M, error) {

	if input == nil {
		input = types.M{}
	}

	input, err := structToMap(input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	if s.memory != nil {
		input = mergeMaps(input, s.memory.All())
	}

	response, err := s.executeLLM(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrLLMExecution, err)
	}

	decodedOutput, err := s.decoder.Decode(response)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	if s.memory != nil {
		err = s.memory.Set(s.namespace, decodedOutput)
		if err != nil {
			return nil, err
		}
	}

	return decodedOutput, nil

}

func (s *Tube) executeLLM(ctx context.Context, input types.M) (string, error) {
	if s.llm.LlmMode == LlmModeCompletion {
		return s.executeLLMCompletion(ctx, input)
	} else if s.llm.LlmMode == LlmModeChat {
		return s.executeLLMChat(ctx, input)
	}

	return "", ErrInvalidLmmMode
}

func (s *Tube) executeLLMCompletion(ctx context.Context, input types.M) (string, error) {
	err := s.llm.Prompt.Format(input)
	if err != nil {
		return "", err
	}

	response, err := s.llm.LlmEngine.Completion(ctx, s.llm.Prompt.String())
	if err != nil {
		return "", err
	}

	return response, nil
}

func (s *Tube) executeLLMChat(ctx context.Context, input types.M) (string, error) {

	for _, promptMessage := range s.llm.Chat.PromptMessages() {
		err := promptMessage.Prompt.Format(input)
		if err != nil {
			return "", err
		}
	}

	response, err := s.llm.LlmEngine.Chat(ctx, s.llm.Chat)
	if err != nil {
		return "", err
	}

	return response, nil
}

// Run chains the steps of the pipeline and returns the output of the last step.
func (p pipeline) Run(ctx context.Context, input types.M) (types.M, error) {
	var err error
	var output types.M
	for i, pipeline := range p.pipes {

		if i == 0 {
			output, err = pipeline.Run(ctx, input)
			if err != nil {
				return nil, err
			}
		} else {
			output, err = pipeline.Run(ctx, output)
			if err != nil {
				return nil, err
			}
		}

	}

	return output, nil
}

func mergeMaps(m1 types.M, m2 types.M) types.M {
	merged := make(types.M)
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}

func structToMap(obj interface{}) (types.M, error) {
	genericMap := types.M{}
	err := mapstructure.Decode(obj, &genericMap)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	return genericMap, nil
}
