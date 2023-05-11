package pipeline

import (
	"context"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
)

const (
	systemPromptTemplate = `You are an helpful assistant. Answer to the questions using only the provided context.
	Don't add any information that is not in the context.
	If you don't know the answer, just say 'I don't know'.`

	userPromptTemplate = "Based on the following context answer to the question.\n\nContext:\n{{.context}}\n\nQuestion: {{.query}}"
)

type QATube struct {
	tube *Tube
}

func NewQATube(llmEngine LlmEngine) *QATube {

	systemPrompt := prompt.New(systemPromptTemplate)
	userPrompt, _ := prompt.NewPromptTemplate(userPromptTemplate, nil)

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

	llm := Llm{
		LlmEngine: llmEngine,
		LlmMode:   LlmModeChat,
		Chat:      chat,
	}

	tube := NewTube(llm)
	return &QATube{
		tube: tube,
	}
}

func (t *QATube) WithPrompt(chat *chat.Chat) *QATube {
	t.tube.llm.Chat = chat
	return t
}

func (t *QATube) Run(ctx context.Context, query string, documents []document.Document) (types.M, error) {

	content := ""
	for _, document := range documents {
		content += document.Content + "\n"
	}

	return t.tube.Run(
		ctx,
		types.M{
			"query":   query,
			"context": content,
		},
	)

}
