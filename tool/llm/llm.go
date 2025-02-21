package llm

import (
	"context"
	"time"

	"github.com/rsest/lingoose/thread"
)

const (
	defaultTimeoutInMinutes = 6
)

type LLM interface {
	Generate(context.Context, *thread.Thread) error
}

type Tool struct {
	llm LLM
}

func New(llm LLM) *Tool {
	return &Tool{
		llm: llm,
	}
}

type Input struct {
	Query string `json:"query" jsonschema:"description=user query"`
}

type Output struct {
	Error  string `json:"error,omitempty"`
	Result string `json:"result,omitempty"`
}

type FnPrototype func(Input) Output

func (t *Tool) Name() string {
	return "llm"
}

func (t *Tool) Description() string {
	return "A tool that uses a language model to generate a response to a user query."
}

func (t *Tool) Fn() any {
	return t.fn
}

//nolint:gosec
func (t *Tool) fn(i Input) Output {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutInMinutes*time.Minute)
	defer cancel()

	th := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(i.Query),
		),
	)

	err := t.llm.Generate(ctx, th)
	if err != nil {
		return Output{Error: err.Error()}
	}

	return Output{Result: th.LastMessage().Contents[0].AsString()}
}
