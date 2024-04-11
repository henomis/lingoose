package gemini

import "fmt"

var (
	ErrGeminiChat   = fmt.Errorf("gemini chat error")
	ErrGeminiNoChat = fmt.Errorf("gemini no chat message recieved")
)

type Model string

func (m Model) String() string {
	return string(m)
}

const (
	Gemini1Pro        Model = "gemini-1.0-pro"
	Gemini1Pro001     Model = "gemini-1.0-pro-001"
	GeminiPro15Latest Model = "gemini-1.5-pro-preview-0409"
)

type StreamCallback func(string)
