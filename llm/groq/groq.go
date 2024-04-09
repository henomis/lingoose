package groq

import (
	"os"

	"github.com/henomis/lingoose/llm/openai"
	goopenai "github.com/sashabaranov/go-openai"
)

const (
	groqAPIEndpoint = "https://api.groq.com/openai/v1"
)

type Groq struct {
	*openai.OpenAI
}

func New() *Groq {
	customConfig := goopenai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	customConfig.BaseURL = groqAPIEndpoint
	customClient := goopenai.NewClientWithConfig(customConfig)

	openaillm := openai.New().WithClient(customClient)
	return &Groq{
		OpenAI: openaillm,
	}
}
