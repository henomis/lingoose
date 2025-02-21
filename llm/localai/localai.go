package localai

import (
	"os"

	"github.com/rsest/lingoose/llm/openai"
	goopenai "github.com/sashabaranov/go-openai"
)

type LocalAI struct {
	*openai.OpenAI
}

func New(endpoint string) *LocalAI {
	customConfig := goopenai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	customConfig.BaseURL = endpoint
	customClient := goopenai.NewClientWithConfig(customConfig)

	openaillm := openai.New().WithClient(customClient)
	openaillm.Name = "localai"
	return &LocalAI{
		OpenAI: openaillm,
	}
}
