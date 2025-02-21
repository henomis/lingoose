package loader

import (
	"context"
	"fmt"
	"os"

	"github.com/rsest/lingoose/document"
	"github.com/rsest/lingoose/types"
	"github.com/sashabaranov/go-openai"
)

type WhisperLoader struct {
	loader Loader

	filename     string
	openAIClient *openai.Client
}

func NewWhisperLoader(filename string) *WhisperLoader {
	openAIApiKey := os.Getenv("OPENAI_API_KEY")

	return &WhisperLoader{
		filename:     filename,
		openAIClient: openai.NewClient(openAIApiKey),
	}
}

func NewWhisper() *WhisperLoader {
	openAIApiKey := os.Getenv("OPENAI_API_KEY")

	return &WhisperLoader{
		openAIClient: openai.NewClient(openAIApiKey),
	}
}

func (w *WhisperLoader) WithClient(client *openai.Client) *WhisperLoader {
	w.openAIClient = client
	return w
}

func (w *WhisperLoader) Load(ctx context.Context) ([]document.Document, error) {
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
		return nil, fmt.Errorf("%w: %w", ErrInternal, err)
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

func (w *WhisperLoader) LoadFromSource(ctx context.Context, source string) ([]document.Document, error) {
	w.filename = source
	return w.Load(ctx)
}
