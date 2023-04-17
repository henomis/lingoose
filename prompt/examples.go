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

func (p *PromptTemplate) AddExamples(examples PromptExamples) error {
	var buffer bytes.Buffer

	buffer.WriteString(examples.Prefix)
	buffer.WriteString(examples.Separator)

	for _, example := range examples.Examples {

		examplePrompt, err := examples.PromptTemplate.Format(Inputs(example))
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
