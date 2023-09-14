package qapipeline

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/index"
	indexoption "github.com/henomis/lingoose/index/option"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
)

const (
	qaTubeSystemPromptTemplate = `You are an helpful assistant. Answer to the questions using only the provided context.
	Don't add any information that is not in the context.
	If you don't know the answer, just say 'I don't know'.`

	qaTubeUserPromptTemplate = "Based on the following context answer to the question.\n\nContext:\n{{.context}}\n\nQuestion: {{.query}}"
)

type Index interface {
	Query(context.Context, string, ...indexoption.Option) (index.SearchResults, error)
}

type QAPipeline struct {
	llmEngine pipeline.LlmEngine
	pipeline  *pipeline.Pipeline
	index     Index
}

func New(llmEngine pipeline.LlmEngine) *QAPipeline {

	systemPrompt := prompt.New(qaTubeSystemPromptTemplate)
	userPrompt := prompt.NewPromptTemplate(qaTubeUserPromptTemplate)

	chat := chat.New(
		chat.PromptMessage{
			Type:   chat.MessageTypeSystem,
			Prompt: systemPrompt,
		},
		chat.PromptMessage{
			Type:   chat.MessageTypeUser,
			Prompt: userPrompt,
		},
	)

	llm := pipeline.Llm{
		LlmEngine: llmEngine,
		LlmMode:   pipeline.LlmModeChat,
		Chat:      chat,
	}

	tube := pipeline.NewTube(llm)
	return &QAPipeline{
		llmEngine: llmEngine,
		pipeline:  pipeline.New(tube),
		index:     nil,
	}
}

func (p *QAPipeline) WithPrompt(chat *chat.Chat) *QAPipeline {
	llm := pipeline.Llm{
		LlmEngine: p.llmEngine,
		LlmMode:   pipeline.LlmModeChat,
		Chat:      chat,
	}

	tube := pipeline.NewTube(llm)

	return &QAPipeline{
		pipeline: pipeline.New(tube),
	}
}

func (q *QAPipeline) WithIndex(index Index) *QAPipeline {
	q.index = index
	return q
}

func (q *QAPipeline) Query(ctx context.Context, query string, opts ...indexoption.Option) (types.M, error) {
	if q.index == nil {
		return nil, fmt.Errorf("retriever is not defined")
	}

	docs, err := q.index.Query(ctx, query, opts...)
	if err != nil {
		return nil, err
	}

	return q.Run(ctx, query, docs.ToDocuments())
}

func (t *QAPipeline) Run(ctx context.Context, query string, documents []document.Document) (types.M, error) {

	content := ""
	for _, document := range documents {
		content += document.Content + "\n"
	}

	return t.pipeline.Run(
		ctx,
		types.M{
			"query":   query,
			"context": content,
		},
	)

}
