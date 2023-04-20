package pipeline

import (
	"github.com/henomis/lingoose/prompt/decoder"
)

type Prompt interface {
	Prompt() string
	Format(input interface{}) error
}

type Llm interface {
	Completion(string) (string, error)
	// Chat(chat chat.Chat) (interface{}, error)
}

type Pipeline struct {
	llm     Llm
	prompt  Prompt
	decoder decoder.Decoder
}

type Pipelines []Pipeline

func New(llm Llm, prompt Prompt, decoder decoder.Decoder) *Pipeline {
	return &Pipeline{
		llm:     llm,
		prompt:  prompt,
		decoder: decoder,
	}
}

func (p *Pipeline) Run(input interface{}) (interface{}, error) {

	err := p.prompt.FormatWithInput(input)
	if err != nil {
		return nil, err
	}

	response, err := p.llm.Completion(p.prompt.Prompt())
	if err != nil {
		return nil, err
	}

	decoded, err := p.decoder(response)
	if err != nil {
		return nil, err
	}

	return decoded, nil

}

func (p Pipelines) Run(input interface{}) (interface{}, error) {
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
