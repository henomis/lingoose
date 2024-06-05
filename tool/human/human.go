package human

import (
	"fmt"
)

type Tool struct {
}

func New() *Tool {
	return &Tool{}
}

type Input struct {
	Question string `json:"question" jsonschema:"description=the question to ask the human"`
}

type Output struct {
	Error  string `json:"error,omitempty"`
	Result string `json:"result,omitempty"`
}

type FnPrototype = func(Input) Output

func (t *Tool) Name() string {
	return "human"
}

func (t *Tool) Description() string {
	return "A tool that asks a question to a human and returns the answer. Use it to interact with a human."
}

func (t *Tool) Fn() any {
	return t.fn
}

func (t *Tool) fn(i Input) Output {
	var answer string

	fmt.Printf("\n\n%s > ", i.Question)
	fmt.Scanln(&answer)

	return Output{Result: answer}
}
