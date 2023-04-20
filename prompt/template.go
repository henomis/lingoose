package prompt

import (
	"bytes"
	"fmt"
	texttemplate "text/template"

	"github.com/mitchellh/mapstructure"
)

type PromptTemplate struct {
	input          interface{}
	template       string
	value          string
	templateEngine *texttemplate.Template
}

func NewPromptTemplate(template string, input interface{}) (*PromptTemplate, error) {

	if input == nil {
		input = map[string]interface{}{}
	}

	genericMap := map[string]interface{}{}
	err := mapstructure.Decode(input, &genericMap)
	if err != nil {
		return nil, ErrDecoding
	}
	input = genericMap

	promptTemplate := &PromptTemplate{
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
func (p *PromptTemplate) Format(input interface{}) error {

	if p.templateEngine == nil {
		err := p.initTemplateEngine()
		if err != nil {
			return ErrTemplateEngine
		}
	}

	if input == nil {
		input = map[string]interface{}{}
	}

	input, err := structToMap(input)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrDecoding, err)
	}

	overallMap := mergeMaps(p.input.(map[string]interface{}), input.(map[string]interface{}))

	var buffer bytes.Buffer
	err = p.templateEngine.Execute(&buffer, overallMap)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrTemplateEngine, err)
	}

	p.value = buffer.String()

	return nil
}

func (p *PromptTemplate) Prompt() string {
	return p.value
}

func (p *PromptTemplate) initTemplateEngine() error {

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
