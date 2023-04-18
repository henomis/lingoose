// Package prompt provides a easy way to format a prompt using the Go template engine.
// Prompts are defined using a template string and a list of inputs.
package prompt

import (
	"bytes"
	"fmt"
	texttemplate "text/template"

	"github.com/go-playground/validator/v10"
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

	if p.Input != nil {
		validate := validator.New()
		if err := validate.Struct(p.Input); err != nil {

			if _, ok := err.(*validator.InvalidValidationError); ok {
				fmt.Println(err)
				return err
			}

			for _, err := range err.(validator.ValidationErrors) {

				fmt.Println(err.Namespace())
				fmt.Println(err.Field())
				fmt.Println(err.StructNamespace())
				fmt.Println(err.StructField())
				fmt.Println(err.Tag())
				fmt.Println(err.ActualTag())
				fmt.Println(err.Kind())
				fmt.Println(err.Type())
				fmt.Println(err.Value())
				fmt.Println(err.Param())
				fmt.Println()
			}

			return err
		}
	}

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
