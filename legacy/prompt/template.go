package prompt

import (
	"bytes"
	"fmt"
	texttemplate "text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/rsest/lingoose/types"
)

type Template struct {
	input          interface{}
	template       string
	value          string
	templateEngine *texttemplate.Template
}

func NewPromptTemplate(text string) *Template {
	promptTemplate := &Template{
		input:    types.M{},
		template: text,
	}

	return promptTemplate
}

func (t *Template) WithInputs(inputs interface{}) *Template {
	t.input = inputs
	return t
}

// Format formats the prompt using the template engine and the provided inputs.
func (t *Template) Format(input types.M) error {
	err := t.initTemplateEngine()
	if err != nil {
		return ErrTemplateEngine
	}

	err = t.decodeInput()
	if err != nil {
		return err
	}

	input, err = structToMap(input)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDecoding, err)
	}

	overallMap := mergeMaps(t.input.(types.M), input)

	var buffer bytes.Buffer
	err = t.templateEngine.Execute(&buffer, overallMap)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTemplateEngine, err)
	}

	t.value = buffer.String()

	return nil
}

func (t *Template) String() string {
	return t.value
}

func (t *Template) initTemplateEngine() error {
	if t.templateEngine != nil {
		return nil
	}

	templateEngine, err := texttemplate.New("prompt").Option("missingkey=zero").Parse(t.template)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTemplateEngine, err)
	}

	t.templateEngine = templateEngine

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

func (t *Template) decodeInput() error {
	genericMap := types.M{}
	err := mapstructure.Decode(t.input, &genericMap)
	if err != nil {
		return ErrDecoding
	}
	t.input = genericMap

	return nil
}
