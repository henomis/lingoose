// Package prompt provides a easy way to format a prompt using the Go template engine.
// Prompts are defined using a template string and a list of inputs.
package prompt

import "errors"

var (
	ErrFormatting     = errors.New("formatting prompt error")
	ErrDecoding       = errors.New("decoding input error")
	ErrTemplateEngine = errors.New("template engine error")
)

type Prompt struct {
	prompt string
}

func New(prompt string) *Prompt {
	return &Prompt{
		prompt: prompt,
	}
}

func (p *Prompt) Format(input interface{}) error {
	return nil
}

func (p *Prompt) Prompt() string {
	return p.prompt
}
