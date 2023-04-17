package prompt

import "bytes"

type PromptExamples struct {
	Examples       []Example
	Separator      string
	PromptTemplate *PromptTemplate
	Prefix         string
	Suffix         string
}

type Example map[string]interface{}

func NewWithExamples(
	inputs []string,
	outputs []string,
	examples PromptExamples,
) (*PromptTemplate, error) {

	var buffer bytes.Buffer

	buffer.WriteString(examples.Prefix)
	buffer.WriteString(examples.Separator)

	for _, example := range examples.Examples {

		examplePrompt, err := examples.PromptTemplate.Format(PromptTemplateInputs(example))
		if err != nil {
			return nil, err
		}

		buffer.WriteString(examplePrompt)
		buffer.WriteString(examples.Separator)
	}

	buffer.WriteString(examples.Suffix)

	return New(
		inputs,
		outputs,
		buffer.String(),
	), nil
}
