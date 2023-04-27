package prompt

import (
	"bytes"
	"fmt"
	texttemplate "text/template"

	"github.com/henomis/lingoose/types"
	"github.com/mitchellh/mapstructure"
)

type Template struct {
	input          interface{}
	template       string
	value          string
	templateEngine *texttemplate.Template
}

func NewPromptTemplate(template string, input interface{}) (*Template, error) {

	if input == nil {
		input = types.M{}
	}

	genericMap := types.M{}
	err := mapstructure.Decode(input, &genericMap)
	if err != nil {
		return nil, ErrDecoding
	}
	input = genericMap

	promptTemplate := &Template{
		input:    input,
		template: template,
	}

	err = promptTemplate.initTemplateEngine()
	if err != nil {
		return nil, ErrTemplateEngine
	}

	return promptTemplate, nil
}

// Format formats the prompt using the template engine and the provided inputs.
func (p *Template) Format(input types.M) error {

	if p.templateEngine == nil {
		err := p.initTemplateEngine()
		if err != nil {
			return ErrTemplateEngine
		}
	}

	if input == nil {
		input = types.M{}
	}

	input, err := structToMap(input)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	overallMap := mergeMaps(p.input.(types.M), input)

	var buffer bytes.Buffer
	err = p.templateEngine.Execute(&buffer, overallMap)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrTemplateEngine, err)
	}

	p.value = buffer.String()

	return nil
}

func (p *Template) Prompt() string {
	return p.value
}

func (p *Template) initTemplateEngine() error {

	if p.templateEngine != nil {
		return nil
	}

	templateEngine, err := texttemplate.New("prompt").Option("missingkey=zero").Parse(p.template)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrTemplateEngine, err)
	}

	p.templateEngine = templateEngine

	return nil
}

func mergeMaps(m1 types.M, m2 types.M) types.M {
	merged := make(types.M)
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = value
	}
	return merged
}

func structToMap(obj interface{}) (types.M, error) {
	genericMap := types.M{}
	err := mapstructure.Decode(obj, &genericMap)
	if err != nil {
		return nil, err
	}

	return genericMap, nil
}
