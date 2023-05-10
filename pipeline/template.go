package pipeline

import (
	"context"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
)

type TubeQa struct {
	tube *Tube
}

func NewQATube(llmEngine LlmEngine) *TubeQa {

	systemPrompt := prompt.New("You are an helpful assistant. Answer to the questions using only " +
		"the provided context. Don't add any information that is not in the context. " +
		"If you don't know the answer, just say 'I don't know'.",
	)
	userPrompt, _ := prompt.NewPromptTemplate(
		"Based on the following context answer to the question.\n\nContext:\n{{.context}}\n\nQuestion: {{.query}}",
		nil,
	)

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
	return &TubeQa{
		tube: tube,
	}
}

func (t *TubeQa) Run(ctx context.Context, query string, similarities []index.SearchResponse) (types.M, error) {

	content := ""
	for _, similarity := range similarities {
		content += similarity.Document.Content + "\n"
	}

	return t.tube.Run(
		ctx,
		types.M{
			"query":   query,
			"context": content,
		},
	)

}
