// Package pipeline provides a way to chain multiple llm executions.
package pipeline

import (
	"context"
	"errors"
	"fmt"

	"github.com/henomis/lingoose/decoder"
	"github.com/mitchellh/mapstructure"
)

var (
	ErrDecoding       = errors.New("decoding input error")
	ErrInvalidLmmMode = errors.New("invalid LLM mode")
)

type Memory interface {
	Get(key string) interface{}
	Set(key string, value interface{}) error
	All() map[string]interface{}
	Delete(key string) error
	Clear() error
}

type Decoder interface {
	Decode(input string) (interface{}, error)
}

type Step struct {
	name    string
	llm     Llm
	decoder Decoder
	memory  Memory
}

type Pipeline struct {
	steps []*Step
}

func New(steps ...*Step) Pipeline {
	return Pipeline{
		steps: steps,
	}
}

func NewStep(
	name string,
	llm Llm,
	outputDecoder Decoder,
	memory Memory,
) *Step {

	if outputDecoder == nil {
		outputDecoder = decoder.NewDefaultDecoder()
	}

	return &Step{
		name:    name,
		llm:     llm,
		decoder: outputDecoder,
		memory:  memory,
	}
}

// Run execute the step and return the output.
// The prompt is formatted with the input and the output of the prompt is used as input for the LLM.
// If the step has a memory, the output is stored in the memory.
func (s *Step) Run(ctx context.Context, input interface{}) (interface{}, error) {

	if input == nil {
		input = map[string]interface{}{}
	}

	input, err := structToMap(input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	if s.memory != nil {
		input = mergeMaps(input.(map[string]interface{}), s.memory.All())
	}

	response, err := s.executeLLM(ctx, input)
	if err != nil {
		return nil, err
	}

	decodedOutput, err := s.decoder.Decode(response)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	if s.memory != nil {
		err = s.memory.Set(s.name, decodedOutput)
		if err != nil {
			return nil, err
		}
	}

	return decodedOutput, nil

}

func (s *Step) executeLLM(ctx context.Context, input interface{}) (string, error) {
	if s.llm.LlmMode == LlmModeCompletion {
		return s.executeLLMCompletion(ctx, input)
	} else if s.llm.LlmMode == LlmModeChat {
		return s.executeLLMChat(ctx, input)
	}

	return "", ErrInvalidLmmMode
}

func (s *Step) executeLLMCompletion(ctx context.Context, input interface{}) (string, error) {
	err := s.llm.Prompt.Format(input)
	if err != nil {
		return "", err
	}

	response, err := s.llm.LlmEngine.Completion(ctx, s.llm.Prompt.Prompt())
	if err != nil {
		return "", err
	}

	return response, nil
}

func (s *Step) executeLLMChat(ctx context.Context, input interface{}) (string, error) {

	for _, promptMessage := range s.llm.Chat.PromptMessages {
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

func (p *Pipeline) AddNext(step *Step) {
	p.steps = append(p.steps, step)
}

// Run chains the steps of the pipeline and returns the output of the last step.
func (p Pipeline) Run(ctx context.Context, input interface{}) (interface{}, error) {
	var err error
	var output interface{}
	for i, pipeline := range p.steps {

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

func mergeMaps(m1 map[string]interface{}, m2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}

func structToMap(obj interface{}) (map[string]interface{}, error) {
	genericMap := map[string]interface{}{}
	mapstructure.Decode(obj, &genericMap)
	return genericMap, nil
}
