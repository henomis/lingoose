package groq

import (
	"os"

	"github.com/rsest/lingoose/llm/openai"
	goopenai "github.com/sashabaranov/go-openai"
)

const (
	groqAPIEndpoint = "https://api.groq.com/openai/v1"
)

type Groq struct {
	*openai.OpenAI
}

func New() *Groq {
	customConfig := goopenai.DefaultConfig(os.Getenv("GROQ_API_KEY"))
	customConfig.BaseURL = groqAPIEndpoint
	customClient := goopenai.NewClientWithConfig(customConfig)

	openaillm := openai.New().WithClient(customClient)
	openaillm.Name = "groq"
	return &Groq{
		OpenAI: openaillm,
	}
}
