package summarize

import (
	"context"

	"github.com/henomis/lingoose/assistant"
	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

type CallbackFn func(t *thread.Thread, i, n int)

type LLM interface {
	Generate(context.Context, *thread.Thread) error
}

type Loader interface {
	Load(ctx context.Context) ([]document.Document, error)
}

type Summarize struct {
	assistant  *assistant.Assistant
	loader     Loader
	callbackFn CallbackFn
}

func New(llm LLM, loader Loader) *Summarize {
	return &Summarize{
		assistant: assistant.New(llm),
		loader:    loader,
	}
}

func (s *Summarize) WithCallback(callbackFn CallbackFn) *Summarize {
	s.callbackFn = callbackFn
	return s
}

func (s *Summarize) Run(ctx context.Context) (*string, error) {
	documents, err := s.loader.Load(ctx)
	if err != nil {
		return nil, err
	}

	nDocuments := len(documents)
	summary := ""

	if s.callbackFn != nil {
		s.callbackFn(s.assistant.Thread(), 0, nDocuments)
	}

	for i, document := range documents {
		prompt := refinePrompt
		if i == 0 {
			prompt = summaryPrompt
		}

		s.assistant.Thread().ClearMessages().AddMessage(
			thread.NewAssistantMessage().AddContent(
				thread.NewTextContent(
					prompt,
				).Format(
					types.M{
						"text":   document.Content,
						"output": summary,
					},
				),
			),
		)

		err := s.assistant.Run(ctx)
		if err != nil {
			return nil, err
		}

		if s.callbackFn != nil {
			s.callbackFn(s.assistant.Thread(), i+1, nDocuments)
		}

		summary = s.assistant.Thread().LastMessage().Contents[0].Data.(string)
	}

	return &summary, nil
}
