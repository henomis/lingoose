package loader

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
	"github.com/sashabaranov/go-openai"
)

type whisperLoader struct {
	loader loader

	filename     string
	openAIClient *openai.Client
}

func NewWhisperLoader(filename string) *whisperLoader {

	openAIApiKey := os.Getenv("OPENAI_API_KEY")

	return &whisperLoader{
		filename:     filename,
		openAIClient: openai.NewClient(openAIApiKey),
	}
}

func (w *whisperLoader) WithClient(client *openai.Client) *whisperLoader {
	w.openAIClient = client
	return w
}

func (w *whisperLoader) Load(ctx context.Context) ([]document.Document, error) {

	err := isFile(w.filename)
	if err != nil {
		return nil, err
	}

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: w.filename,
	}
	resp, err := w.openAIClient.CreateTranscription(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	documents := []document.Document{
		{
			Content: resp.Text,
			Metadata: types.Meta{
				SourceMetadataKey: w.filename,
			},
		},
	}

	if w.loader.textSplitter != nil {
		documents = w.loader.textSplitter.SplitDocuments(documents)
	}

	return documents, nil
}
