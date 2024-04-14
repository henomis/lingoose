package pipeline

import (
	"context"

	"github.com/henomis/lingoose/legacy/chat"
	"github.com/henomis/lingoose/types"
)

type Llm struct {
	LlmEngine LlmEngine
	LlmMode   LlmMode
	Prompt    Prompt
	Chat      *chat.Chat
}

type LlmMode int

const (
	LlmModeChat LlmMode = iota
	LlmModeCompletion
)

type Prompt interface {
	String() string
	Format(input types.M) error
}

type LlmEngine interface {
	Completion(ctx context.Context, prompt string) (string, error)
	Chat(ctx context.Context, chat *chat.Chat) (string, error)
}
