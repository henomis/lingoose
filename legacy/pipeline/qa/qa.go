package qapipeline

import (
	"context"
	"fmt"

	"github.com/rsest/lingoose/document"
	"github.com/rsest/lingoose/index"
	indexoption "github.com/rsest/lingoose/index/option"
	"github.com/rsest/lingoose/legacy/chat"
	"github.com/rsest/lingoose/legacy/pipeline"
	"github.com/rsest/lingoose/legacy/prompt"
	"github.com/rsest/lingoose/types"
)

const (
	qaTubeSystemPromptTemplate = `You are an helpful assistant. Answer to the questions using only the provided context.
	Don't add any information that is not in the context.
	If you don't know the answer, just say 'I don't know'.`

	//nolint:lll
	qaTubeUserPromptTemplate = "Based on the following context answer to the question.\n\nContext:\n{{.context}}\n\nQuestion: {{.query}}"

	qaTubePromptTemplate = "Context information is below\n--------------------\n{{.context}}\n" +
		"--------------------\nGiven the context information and not prior knowledge, answer the following question.\n" +
		"Question: {{.query}}\nAnswer:"
	qaTubePromptRefineTemplate = "The original question is as follows: {{.query}}\n" +
		"We have provided an existing answer: {{.answer}}\n" +
		"We have the opportunity to refine the existing answer (only if needed) with some more context below.\n" +
		"----------------------\n{{.context}}\n----------------------\n" +
		"Given the new context, refine the original answer to better answer the question.\n" +
		"If the context isn't useful, return the original answer.\n" +
		"Refined answer:"
)

type Mode string

const (
	ModeSimple Mode = "simple"
	ModeRefine Mode = "refine"
)

type Index interface {
	Query(context.Context, string, ...indexoption.Option) (index.SearchResults, error)
}

type QAPipeline struct {
	llmEngine pipeline.LlmEngine
	pipeline  *pipeline.Pipeline
	mode      Mode
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
		mode:      ModeSimple,
	}
}

func (q *QAPipeline) WithPrompt(chat *chat.Chat) *QAPipeline {
	llm := pipeline.Llm{
		LlmEngine: q.llmEngine,
		LlmMode:   pipeline.LlmModeChat,
		Chat:      chat,
	}

	tube := pipeline.NewTube(llm)

	return &QAPipeline{
		llmEngine: q.llmEngine,
		pipeline:  pipeline.New(tube),
		index:     q.index,
		mode:      q.mode,
	}
}

func (q *QAPipeline) WithIndex(index Index) *QAPipeline {
	q.index = index
	return q
}

func (q *QAPipeline) WithMode(mode Mode) *QAPipeline {
	q.mode = mode
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

func (q *QAPipeline) Run(ctx context.Context, query string, documents []document.Document) (types.M, error) {
	content := ""
	for _, document := range documents {
		content += document.Content + "\n"
	}

	if q.mode == ModeSimple {
		return q.pipeline.Run(
			ctx,
			types.M{
				"query":   query,
				"context": content,
			},
		)
	}

	return q.runRefine(ctx, query, documents)
}

func (q *QAPipeline) runRefine(ctx context.Context, query string, documents []document.Document) (types.M, error) {
	var currentResponse string
	var output types.M
	var err error

	for i, document := range documents {
		context := document.Content

		var qaPrompt *prompt.Template
		if i == 0 {
			qaPrompt = prompt.NewPromptTemplate(qaTubePromptTemplate)
		} else {
			qaPrompt = prompt.NewPromptTemplate(qaTubePromptRefineTemplate)
		}

		llm := pipeline.Llm{
			LlmEngine: q.llmEngine,
			LlmMode:   pipeline.LlmModeCompletion,
			Prompt:    qaPrompt,
		}

		tube := pipeline.NewTube(llm)
		q.pipeline = pipeline.New(tube)
		output, err = q.pipeline.Run(
			ctx, types.M{
				"query":   query,
				"answer":  currentResponse,
				"context": context,
			},
		)
		if err != nil {
			return nil, err
		}

		response, ok := output[types.DefaultOutputKey].(string)
		if !ok {
			return nil, fmt.Errorf("invalid response type")
		}
		currentResponse = response
	}

	return output, nil
}
