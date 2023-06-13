package prompt

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/types"
	"github.com/sashabaranov/go-openai"
)

type AudioResponseFormat string

const (
	AudioResponseFormatText AudioResponseFormat = ""
	AudioResponseFormatJSON AudioResponseFormat = "json"
	AudioResponseFormatSRT  AudioResponseFormat = "srt"
	AudioResponseFormatVTT  AudioResponseFormat = "vtt"
)

type whisperPrompt struct {
	openAIClient        *openai.Client
	filePath            string
	ctx                 context.Context
	audioResponseFormat AudioResponseFormat
}

func NewPromptFromAudioFile(ctx context.Context, filePath string, audioResponseFormat AudioResponseFormat) (*whisperPrompt, error) {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	return &whisperPrompt{
		openAIClient:        openai.NewClient(openAIKey),
		filePath:            filePath,
		ctx:                 ctx,
		audioResponseFormat: audioResponseFormat,
	}, nil
}

func (p *whisperPrompt) WithClient(client *openai.Client) *whisperPrompt {
	p.openAIClient = client
	return p
}

func (p *whisperPrompt) Format(input types.M) error {
	return nil
}

func (p *whisperPrompt) String() string {
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: p.filePath,
		Format:   openai.AudioResponseFormat(p.audioResponseFormat),
	}
	resp, err := p.openAIClient.CreateTranscription(p.ctx, req)
	if err != nil {
		return ""
	}

	return resp.Text
}
