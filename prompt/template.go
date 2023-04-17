package prompt

import (
	"bytes"
	"fmt"
	"text/template"
)

type Inputs map[string]interface{}
type PromptTemplateOutputs map[string]interface{}

type PromptTemplate struct {
	inputs   []string
	outputs  []string
	template string
	partials *Inputs

	inputsSet      map[string]struct{}
	templateEngine *template.Template
}

func NewPromptTemplate(inputsList []string, outputsList []string, template string, partials *Inputs) *PromptTemplate {

	return &PromptTemplate{
		inputs:   inputsList,
		outputs:  outputsList,
		template: template,
		partials: partials,

		inputsSet: buildInputsSet(inputsList),
	}
}

func NewPromptTemplateWithExamples(inputsList []string, outputsList []string, examples PromptExamples) (*PromptTemplate, error) {

	promptTemplate := NewPromptTemplate(inputsList, outputsList, "", nil)

	err := promptTemplate.AddExamples(examples)
	if err != nil {
		return nil, err
	}

	return promptTemplate, nil
}

func NewPromptTemplateFromLangchain(url string) (*PromptTemplate, error) {

	var langchainPromptTemplate langchainPromptTemplate
	if err := langchainPromptTemplate.ImportFromLangchain(url); err != nil {
		return nil, err
	}

	return langchainPromptTemplate.toPromptTemplate(), nil
}

func (p *PromptTemplate) SetPartials(partials *Inputs) {
	p.partials = partials
}

func (p *PromptTemplate) Format(promptTemplateInputs Inputs) (string, error) {

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

// ValidateInputs checks if some inputs do not match the inputsSet
func (p *PromptTemplate) validateInputs(promptTemplateInputs Inputs) error {

	for input := range promptTemplateInputs {
		if _, ok := p.inputsSet[input]; !ok {
			return fmt.Errorf("invalid input %s", input)
		}
	}

	return nil
}

func buildInputsSet(inputs []string) map[string]struct{} {
	inputsSet := make(map[string]struct{})
	for _, input := range inputs {
		inputsSet[input] = struct{}{}
	}
	return inputsSet
}
