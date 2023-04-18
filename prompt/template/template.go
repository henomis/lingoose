package template

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/henomis/lingopipes/prompt/example"
	"github.com/henomis/lingopipes/prompt/langchain"
)

type Inputs map[string]interface{}
type Outputs map[string]interface{}

type Template struct {
	inputs   []string
	outputs  []string
	template string
	partials *Inputs

	inputsSet      map[string]struct{}
	templateEngine *template.Template
}

func New(inputsList []string, outputsList []string, template string, partials *Inputs) *Template {

	return &Template{
		inputs:   inputsList,
		outputs:  outputsList,
		template: template,
		partials: partials,

		inputsSet: buildInputsSet(inputsList),
	}
}

func NewWithExamples(
	inputsList []string,
	outputsList []string,
	examples example.Examples,
	exampleTemplate *Template,
) (*Template, error) {

	promptTemplate := New(inputsList, outputsList, "", nil)

	err := promptTemplate.addExamples(examples, exampleTemplate)
	if err != nil {
		return nil, err
	}

	return promptTemplate, nil
}

func NewFromLangchain(url string) (*Template, error) {

	promptTemplate, err := langchain.New(url)
	if err != nil {
		return nil, err
	}

	return New(promptTemplate.InputVariables, []string{}, promptTemplate.ConvertedTemplate(), nil), nil
}

func (p *Template) SetPartials(partials *Inputs) {
	p.partials = partials
}

func (p *Template) Format(promptTemplateInputs Inputs) (string, error) {

	if err := p.validateInputs(promptTemplateInputs); err != nil {
		return "", err
	}

	// add partials to inputs
	if p.partials != nil {
		for key, value := range *p.partials {
			promptTemplateInputs[key] = value
		}
	}

	p.templateEngine = template.Must(template.New("prompt").Parse(p.template))

	var output bytes.Buffer
	err := p.templateEngine.Execute(&output, promptTemplateInputs)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func (p *Template) InputsSet() map[string]struct{} {
	return p.inputsSet
}

// ValidateInputs checks if some inputs do not match the inputsSet
func (p *Template) validateInputs(promptTemplateInputs Inputs) error {

	for input := range promptTemplateInputs {
		if _, ok := p.inputsSet[input]; !ok {
			return fmt.Errorf("invalid input %s", input)
		}
	}

	return nil
}

func (p *Template) addExamples(examples example.Examples, examplesTemplate *Template) error {
	var buffer bytes.Buffer

	buffer.WriteString(examples.Prefix)
	buffer.WriteString(examples.Separator)

	for _, example := range examples.Examples {

		examplePrompt, err := examplesTemplate.Format(Inputs(example))
		if err != nil {
			return err
		}

		buffer.WriteString(examplePrompt)
		buffer.WriteString(examples.Separator)
	}

	buffer.WriteString(examples.Suffix)

	p.template = buffer.String()

	return nil
}

func buildInputsSet(inputs []string) map[string]struct{} {
	inputsSet := make(map[string]struct{})
	for _, input := range inputs {
		inputsSet[input] = struct{}{}
	}
	return inputsSet
}
