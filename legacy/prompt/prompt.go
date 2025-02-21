// Package prompt provides a easy way to format a prompt using the Go template engine.
// Prompts are defined using a template string and a list of inputs.
package prompt

import (
	"errors"

	"github.com/rsest/lingoose/types"
)

var (
	ErrFormatting     = errors.New("formatting prompt error")
	ErrDecoding       = errors.New("decoding input error")
	ErrTemplateEngine = errors.New("template engine error")
)

type Prompt struct {
	prompt string
}

func New(text string) *Prompt {
	return &Prompt{
		prompt: text,
	}
}

func (p *Prompt) Format(input types.M) error {
	_ = input
	return nil
}

func (p *Prompt) String() string {
	return p.prompt
}
