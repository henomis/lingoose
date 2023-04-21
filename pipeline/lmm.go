package pipeline

import "github.com/henomis/lingoose/chat"

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
	Completion(string) (string, error)
	Chat(chat *chat.Chat) (string, error)
}
