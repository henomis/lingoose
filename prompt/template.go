package prompt

import (
	"bytes"
	texttemplate "text/template"
)

type PromptTemplate struct {
	Input    interface{}
	Template string

	value          string
	templateEngine *texttemplate.Template
}

func NewPromptTemplate(template string, input interface{}) *PromptTemplate {
	return &PromptTemplate{
		Input:    input,
		Template: template,
	}
}

// Format formats the prompt using the template engine and the provided inputs.
func (p *PromptTemplate) Format() error {

	if p.Input == nil {
		return nil
	}

	err := p.initTemplateEngine()
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	err = p.templateEngine.Execute(&buffer, p.Input)
	if err != nil {
		return err
	}

	p.value = buffer.String()

	return nil
}

// Format formats the prompt using the template engine and the provided inputs.
func (p *PromptTemplate) FormatWithInput(input interface{}) error {

	if input == nil {
		return nil
	}

	err := p.initTemplateEngine()
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	err = p.templateEngine.Execute(&buffer, input)
	if err != nil {
		return err
	}

	p.value = buffer.String()

	return nil
}

func (p *PromptTemplate) initTemplateEngine() error {

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

func (p *PromptTemplate) Prompt() string {
	return p.value
}
