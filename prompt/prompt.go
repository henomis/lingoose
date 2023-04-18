// Package prompt provides a easy way to format a prompt using the Go template engine.
// Prompts are defined using a template string and a list of inputs.
package prompt

import (
	"bytes"
	texttemplate "text/template"
)

type OudputDecoder interface {
	Decode(interface{}) error
}

type OutputDecoderFn func(string) OudputDecoder

type Prompt struct {
	Input         interface{}
	Output        interface{}
	OutputDecoder OutputDecoderFn
	Template      string

	templateEngine *texttemplate.Template
}

func New(input interface{}, outputHandler OutputDecoderFn, template string) (*Prompt, error) {
	return &Prompt{
		Input:         input,
		OutputDecoder: outputHandler,
		Template:      template,
	}, nil
}

// Format formats the prompt using the template engine and the provided inputs.
func (p *Prompt) Format() (string, error) {

	if p.Input == nil {
		return p.Template, nil
	}

	err := p.init()
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	err = p.templateEngine.Execute(&output, p.Input)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func (p *Prompt) init() error {

	if p.templateEngine != nil {
		return nil
	}

	templateEngine, err := texttemplate.New("prompt").Parse(p.Template)
	if err != nil {
		return err
	}

	p.templateEngine = templateEngine

	return nil
}
