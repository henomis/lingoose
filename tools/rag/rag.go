package rag

import (
	"context"
	"strings"

	"github.com/henomis/lingoose/rag"
)

type Tool struct {
	rag   *rag.RAG
	topic string
}

func New(rag *rag.RAG, topic string) *Tool {
	return &Tool{
		rag:   rag,
		topic: topic,
	}
}

type Input struct {
	Query string `json:"rag_query" jsonschema:"description=search query"`
}

type Output struct {
	Error  string `json:"error,omitempty"`
	Result string `json:"result,omitempty"`
}

type FnPrototype = func(Input) Output

func (t *Tool) Name() string {
	return "rag"
}

func (t *Tool) Description() string {
	return "A tool that searches information ONLY for this topic: " + t.topic + ". DO NOT use this tool for other topics."
}

func (t *Tool) Fn() any {
	return t.fn
}

//nolint:gosec
func (t *Tool) fn(i Input) Output {
	results, err := t.rag.Retrieve(context.Background(), i.Query)
	if err != nil {
		return Output{Error: err.Error()}
	}

	// Return the output as a string.
	return Output{Result: strings.Join(results, "\n")}
}
