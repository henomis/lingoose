package pipeline

import (
	"context"

	"github.com/henomis/lingoose/chat"
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
	Prompt() string
	Format(input interface{}) error
}

type LlmEngine interface {
	Completion(ctx context.Context, prompt string) (string, error)
	Chat(ctx context.Context, chat *chat.Chat) (string, error)
}
