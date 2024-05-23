package toolrouter

import (
	"context"

	"github.com/henomis/lingoose/thread"
)

type TTool interface {
	Description() string
	Name() string
	Fn() any
}

type Tool struct {
	llm   LLM
	tools []TTool
}

type LLM interface {
	Generate(context.Context, *thread.Thread) error
}

func New(llm LLM, tools ...TTool) *Tool {
	return &Tool{
		tools: tools,
		llm:   llm,
	}
}

type Input struct {
	Query string `json:"query" jsonschema:"description=user query"`
}

type Output struct {
	Error  string `json:"error,omitempty"`
	Result any    `json:"result,omitempty"`
}

type FnPrototype func(Input) Output

func (t *Tool) Name() string {
	return "query_router"
}

func (t *Tool) Description() string {
	return "A tool that select the right tool to answer to user queries."
}

func (t *Tool) Fn() any {
	return t.fn
}

//nolint:gosec
func (t *Tool) fn(i Input) Output {
	query := "Here's a list of available tools:\n\n"
	for _, tool := range t.tools {
		query += "Name: " + tool.Name() + "\nDescription: " + tool.Description() + "\n\n"
	}

	query += "\nPlease select the right tool that can better answer the query '" + i.Query +
		"'. Give me only the name of the tool, nothing else."

	th := thread.New().AddMessage(
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(query),
		),
	)

	err := t.llm.Generate(context.Background(), th)
	if err != nil {
		return Output{Error: err.Error()}
	}

	return Output{Result: th.LastMessage().Contents[0].AsString()}
}
