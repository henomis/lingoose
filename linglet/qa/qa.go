package qa

import (
	"context"

	"github.com/henomis/lingoose/assistant"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/rag"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

const (
	defaultTopK = 3
)

type LLM interface {
	Generate(context.Context, *thread.Thread) error
}

type QA struct {
	llm            LLM
	index          *index.Index
	subDocumentRAG *rag.SubDocumentRAG
}

func New(
	llm LLM,
	index *index.Index,
) *QA {
	subDocumentRAG := rag.NewSubDocument(index, llm).WithTopK(defaultTopK)

	return &QA{
		llm:            llm,
		index:          index,
		subDocumentRAG: subDocumentRAG,
	}
}

func (qa *QA) refinePrompt(ctx context.Context, prompt string) (string, error) {
	t := thread.New().AddMessage(
		thread.NewAssistantMessage().AddContent(
			thread.NewTextContent(refinementPrompt).
				Format(
					types.M{
						"prompt": prompt,
					},
				),
		),
	)

	err := qa.llm.Generate(ctx, t)
	if err != nil {
		return prompt, err
	}

	return t.LastMessage().Contents[0].AsString(), nil
}

func (qa *QA) AddSource(ctx context.Context, source string) error {
	return qa.subDocumentRAG.AddSources(ctx, source)
}

func (qa *QA) Run(ctx context.Context, prompt string) (string, error) {
	refinedPromt, err := qa.refinePrompt(ctx, prompt)
	if err != nil {
		return "", err
	}

	a := assistant.New(
		qa.llm,
	).WithParameters(
		assistant.Parameters{
			AssistantName:      "AI Assistant",
			AssistantIdentity:  "a helpful and polite assistant",
			AssistantScope:     "to answer questions",
			CompanyName:        "",
			CompanyDescription: "",
		},
	).WithRAG(qa.subDocumentRAG).WithThread(
		thread.New().AddMessages(
			thread.NewUserMessage().AddContent(
				thread.NewTextContent(refinedPromt),
			),
		),
	)

	err = a.Run(ctx)
	if err != nil {
		return "", err
	}

	return a.Thread().LastMessage().Contents[0].AsString(), nil
}
