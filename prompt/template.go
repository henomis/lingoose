package prompt

import (
	"bytes"
	"fmt"
	"text/template"
)

type Inputs map[string]interface{}
type PromptTemplateOutputs map[string]interface{}

type PromptTemplate struct {
	Inputs   []string `json:"inputs" yaml:"inputs"`
	Outputs  []string `json:"outputs" yaml:"outputs"`
	Template string   `json:"template" yaml:"template"`

	inputsSet map[string]struct{}
	template  *template.Template
}

func New(
	inputsList []string,
	outputsList []string,
	template string,
) *PromptTemplate {

	return &PromptTemplate{
		Inputs:   inputsList,
		Outputs:  outputsList,
		Template: template,

		inputsSet: buildInputsSet(inputsList),
	}
}

func NewFromLangchain(url string) (*PromptTemplate, error) {

	var langchainPromptTemplate langchainPromptTemplate
	if err := langchainPromptTemplate.ImportFromLangchain(url); err != nil {
		return nil, err
	}

	return langchainPromptTemplate.toPromptTemplate(), nil
}

func (p *PromptTemplate) Format(promptTemplateInputs Inputs) (string, error) {

	if err := p.validateInputs(promptTemplateInputs); err != nil {
		return "", err
	}

	p.template = template.Must(template.New("prompt").Parse(p.Template))

	var output bytes.Buffer
	err := p.template.Execute(&output, promptTemplateInputs)
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
