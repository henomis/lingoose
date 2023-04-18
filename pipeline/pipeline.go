package pipeline

import (
	"github.com/henomis/lingoose/llm"
	"github.com/henomis/lingoose/prompt"
)

type Pipeline struct {
	llm    llm.Llm
	prompt *prompt.Prompt
}

type Pipelines []Pipeline

func New(llm llm.Llm, prompt *prompt.Prompt) *Pipeline {
	return &Pipeline{
		llm:    llm,
		prompt: prompt,
	}
}

func (p *Pipeline) Run() (interface{}, error) {
	response, err := p.llm.Completion(p.prompt)
	if err != nil {
		return nil, err
	}
	_ = response

	return p.prompt.Output, nil
}

func (p Pipelines) Run() (interface{}, error) {
	var err error
	var promptTemplate *prompt.Prompt

	for i, pipeline := range p {

		if i == 0 {
			promptTemplate = pipeline.prompt
		} else {
			promptTemplate, err = prompt.New(
				p[i-1].prompt.Output,
				pipeline.prompt.Output,
				pipeline.prompt.OutputDecoder,
				pipeline.prompt.Template,
			)
			if err != nil {
				return nil, err
			}
			p[i].prompt = promptTemplate
		}

		_, err := pipeline.llm.Completion(promptTemplate)
		if err != nil {
			return nil, err
		}

	}

	return p[len(p)-1].prompt.Output, nil
}
