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

type WhisperPrompt struct {
	openAIClient        *openai.Client
	filePath            string
	ctx                 context.Context
	audioResponseFormat AudioResponseFormat
}

func NewPromptFromAudioFile(
	ctx context.Context,
	filePath string,
	audioResponseFormat AudioResponseFormat,
) (*WhisperPrompt, error) {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	return &WhisperPrompt{
		openAIClient:        openai.NewClient(openAIKey),
		filePath:            filePath,
		ctx:                 ctx,
		audioResponseFormat: audioResponseFormat,
	}, nil
}

func (p *WhisperPrompt) WithClient(client *openai.Client) *WhisperPrompt {
	p.openAIClient = client
	return p
}

func (p *WhisperPrompt) Format(input types.M) error {
	_ = input
	return nil
}

func (p *WhisperPrompt) String() string {
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
