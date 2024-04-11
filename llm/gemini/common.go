package gemini

import "fmt"

var (
	ErrGeminiChat = fmt.Errorf("gemini chat error")
)

type Model string

const (
	Gemini1Pro        Model = "gemini-1.0-pro"
	Gemini1Pro001     Model = "gemini-1.0-pro-001"
	GeminiPro15Latest Model = "gemini-1.5-pro-latest"
)

type StreamCallback func(string)
