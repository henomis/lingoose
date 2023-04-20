// Package prompt provides a easy way to format a prompt using the Go template engine.
// Prompts are defined using a template string and a list of inputs.
package prompt

type SimplePrompt struct {
	prompt string
}

func New(prompt string) *SimplePrompt {
	return &SimplePrompt{
		prompt: prompt,
	}
}

func (p *SimplePrompt) Format() error {
	return nil
}

func (p *SimplePrompt) FormatWithInput(input interface{}) error {
	return nil
}

func (p *SimplePrompt) Prompt() string {
	return p.prompt
}
