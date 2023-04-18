// Package prompt provides a easy way to format a prompt using the Go template engine.
// Prompts are defined using a template string and a list of inputs.
package prompt

import (
	"bytes"
	texttemplate "text/template"

	"github.com/henomis/lingoose/prompt/decoder"
)

type Prompt struct {
	Input         interface{}
	Output        interface{}
	OutputDecoder decoder.DecoderFn
	Template      *string

	templateEngine *texttemplate.Template
}

func New(input interface{}, output interface{}, outputHandler decoder.DecoderFn, template *string) (*Prompt, error) {
	return &Prompt{
		Input:         input,
		Output:        output,
		OutputDecoder: outputHandler,
		Template:      template,
	}, nil
}

// Format formats the prompt using the template engine and the provided inputs.
func (p *Prompt) Format() (string, error) {

	// If the input is a string and there is no template, return the input as is.
	if _, ok := p.Input.(*string); ok && (p.Template == nil) {
		return *p.Input.(*string), nil
	}

	if _, ok := p.Input.(string); ok && (p.Template == nil) {
		return p.Input.(string), nil
	}

	if p.Input == nil {
		return *p.Template, nil
	}

	err := p.initTemplateEngine()
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	err = p.templateEngine.Execute(&buffer, p.Input)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (p *Prompt) initTemplateEngine() error {

	if p.templateEngine != nil {
		return nil
	}

	templateEngine, err := texttemplate.New("prompt").Parse(*p.Template)
	if err != nil {
		return err
	}

	p.templateEngine = templateEngine

	return nil
}
