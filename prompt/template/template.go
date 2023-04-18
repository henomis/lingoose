// Package template provides a easy way to format a prompt using the Go template engine.
// Prompts are defined using a template string and a list of inputs.
package template

import (
	"bytes"
	"text/template"
	texttemplate "text/template"

	"github.com/go-playground/validator/v10"
)

type Decoder interface {
	Decode(interface{}) error
}

type OutputDecoder func(string) Decoder

type Prompt struct {
	Input         interface{}
	Output        interface{}
	OutputDecoder OutputDecoder
	Template      string

	templateEngine *template.Template
	validate       *validator.Validate
}

func New(input interface{}, outputHandler OutputDecoder, template string) (*Prompt, error) {

	templateEngine, err := texttemplate.New("prompt").Parse(template)
	if err != nil {
		return nil, err
	}

	return &Prompt{
		Input:          input,
		OutputDecoder:  outputHandler,
		Template:       template,
		templateEngine: templateEngine,
	}, nil
}

// Format formats the prompt using the template engine and the provided inputs.
func (p *Prompt) Format() (string, error) {

	var output bytes.Buffer
	err := p.templateEngine.Execute(&output, p.Input)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
