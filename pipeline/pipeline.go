package pipeline

import (
	"github.com/henomis/lingoose/prompt/decoder"
	promptdecoder "github.com/henomis/lingoose/prompt/decoder"
	"github.com/mitchellh/mapstructure"
)

type Prompt interface {
	Prompt() string
	Format(input interface{}) error
}

type Llm interface {
	Completion(string) (string, error)
	// Chat(chat chat.Chat) (interface{}, error)
}

type Memory interface {
	Get(key string) interface{}
	Set(key string, value interface{}) error
	All() map[string]interface{}
}

type PipelineStep struct {
	name    string
	llm     Llm
	prompt  Prompt
	output  interface{}
	decoder decoder.Decoder
	memory  Memory
}

type Pipeline []*PipelineStep

func New(steps ...*PipelineStep) Pipeline {
	return steps
}

func NewStep(
	name string,
	llm Llm,
	prompt Prompt,
	output interface{},
	decoder decoder.Decoder,
	memory Memory,
) *PipelineStep {

	if decoder == nil {
		decoder = promptdecoder.NewDefaultDecoder()
	}

	if output == nil {
		output = map[string]interface{}{}
	}

	return &PipelineStep{
		name:    name,
		llm:     llm,
		prompt:  prompt,
		output:  output,
		decoder: decoder,
		memory:  memory,
	}
}

func (p *PipelineStep) Run(input interface{}) (interface{}, error) {

	if input == nil {
		input = map[string]interface{}{}
	}

	input, err := structToMap(input)
	if err != nil {
		return nil, err
	}

	if p.memory != nil {
		input = mergeMaps(input.(map[string]interface{}), p.memory.All())
	}

	err = p.prompt.Format(input)
	if err != nil {
		return nil, err
	}

	response, err := p.llm.Completion(p.prompt.Prompt())
	if err != nil {
		return nil, err
	}

	p.output, err = p.decoder(response, p.output)
	if err != nil {
		return nil, err
	}

	if p.memory != nil {
		err = p.memory.Set(p.name, p.output)
		if err != nil {
			return nil, err
		}
	}

	return p.output, nil

}

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
