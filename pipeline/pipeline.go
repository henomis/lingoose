package pipeline

import (
	"github.com/henomis/lingoose/prompt/decoder"
	promptdecoder "github.com/henomis/lingoose/prompt/decoder"
)

type Prompt interface {
	Prompt() string
	Format(input interface{}) error
}

type Llm interface {
	Completion(string) (string, error)
	// Chat(chat chat.Chat) (interface{}, error)
}

type PipelineStep struct {
	llm     Llm
	prompt  Prompt
	output  interface{}
	decoder decoder.Decoder
}

type Pipeline []*PipelineStep

func New(steps ...*PipelineStep) Pipeline {
	return steps
}

func NewStep(llm Llm, prompt Prompt, output interface{}, decoder decoder.Decoder) *PipelineStep {

	if decoder == nil {
		decoder = promptdecoder.NewDefaultDecoder()
	}

	if output == nil {
		output = map[string]interface{}{}
	}

	return &PipelineStep{
		llm:     llm,
		prompt:  prompt,
		output:  output,
		decoder: decoder,
	}
}

func (p *PipelineStep) Run(input interface{}) (interface{}, error) {

	err := p.prompt.Format(input)
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
