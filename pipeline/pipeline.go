package pipeline

import (
	"errors"
	"fmt"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/decoder"
	"github.com/mitchellh/mapstructure"
)

var (
	ErrDecoding = errors.New("decoding input error")
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

type Pipeline []*Step

func New(steps ...*Step) Pipeline {
	return steps
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
func (p *Step) Run(input interface{}) (interface{}, error) {

	if input == nil {
		input = map[string]interface{}{}
	}

	input, err := structToMap(input)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	if p.memory != nil {
		input = mergeMaps(input.(map[string]interface{}), p.memory.All())
	}

	err = p.llm.Prompt.Format(input)
	if err != nil {
		return nil, err
	}

	response, err := p.llm.LlmEngine.Completion(p.llm.Prompt.Prompt())
	if err != nil {
		return nil, err
	}

	decodedOutput, err := p.decoder.Decode(response)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	if p.memory != nil {
		err = p.memory.Set(p.name, decodedOutput)
		if err != nil {
			return nil, err
		}
	}

	return decodedOutput, nil

}

func (p *Step) executeLLM(input interface{}) (interface{}, error) {
	if p.llm.LlmMode == LlmModeChat {
		return p.executeLLMCompletion(input)
	} else if p.llm.LlmMode == LlmModeCompletion {
		return p.executeLLMChat(input)
	}

	return nil, errors.New("invalid LLM mode")
}

func (p *Step) executeLLMCompletion(input interface{}) (string, error) {
	response, err := p.llm.LlmEngine.Completion(p.llm.Prompt.Prompt())
	if err != nil {
		return "", err
	}

	return response, nil
}

func (p *Step) executeLLMChat(input interface{}) (interface{}, error) {

	messages, err := p.llm.LlmEngine.Chat(input.(*chat.Chat))
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// Run chains the steps of the pipeline and returns the output of the last step.
func (p Pipeline) Run(input interface{}) (interface{}, error) {
	var err error
	var output interface{}
	for i, pipeline := range p {

		if i == 0 {
			output, err = pipeline.Run(input)
			if err != nil {
				return nil, err
			}
		} else {
			output, err = pipeline.Run(output)
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
