package prompt

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/types"
	"github.com/sashabaranov/go-openai"
)

type whisperPrompt struct {
	openAIClient *openai.Client
	filePath     string
	ctx          context.Context
}

func NewPromptFromAudioFile(ctx context.Context, filePath string) (*whisperPrompt, error) {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	return &whisperPrompt{
		openAIClient: openai.NewClient(openAIKey),
		filePath:     filePath,
		ctx:          ctx,
	}, nil
}

func (p *whisperPrompt) Format(input types.M) error {
	return nil
}

func (p *whisperPrompt) String() string {
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: p.filePath,
	}
	resp, err := p.openAIClient.CreateTranscription(p.ctx, req)
	if err != nil {
		return ""
	}

	return resp.Text
}
