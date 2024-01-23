package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
)

const (
	hfDefaultSpeechRecognitionModel = "facebook/wav2vec2-large-960h-lv60-self"
)

type HFSpeechRecognition struct {
	loader Loader

	mediaFile string
	token     string
	model     string
}

type hfSpeechRecognitionResponse struct {
	Text string `json:"text,omitempty"`
}

func NewHFSpeechRecognitionLoader(mediaFile string) *HFSpeechRecognition {
	return &HFSpeechRecognition{
		mediaFile: mediaFile,
		model:     hfDefaultSpeechRecognitionModel,
		token:     os.Getenv("HUGGING_FACE_HUB_TOKEN"),
	}
}

func (h *HFSpeechRecognition) WithToken(token string) *HFSpeechRecognition {
	h.token = token
	return h
}

func (h *HFSpeechRecognition) WithModel(model string) *HFSpeechRecognition {
	h.model = model
	return h
}

func (h *HFSpeechRecognition) WithTextSplitter(textSplitter TextSplitter) *HFSpeechRecognition {
	h.loader.textSplitter = textSplitter
	return h
}

func (h *HFSpeechRecognition) Load(ctx context.Context) ([]document.Document, error) {
	err := isFile(h.mediaFile)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInternal, err)
	}

	responseBytes, err := hfMediaHTTPCall(ctx, h.token, h.model, h.mediaFile)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInternal, err)
	}

	response := hfSpeechRecognitionResponse{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInternal, err)
	}

	var documents []document.Document
	document := document.Document{
		Content: response.Text,
		Metadata: types.Meta{
			"source": h.mediaFile,
		},
	}

	document.Content = strings.TrimSpace(document.Content)
	documents = append(documents, document)

	if h.loader.textSplitter != nil {
		documents = h.loader.textSplitter.SplitDocuments(documents)
	}

	return documents, nil
}

func (h *HFSpeechRecognition) LoadFromSource(ctx context.Context, source string) ([]document.Document, error) {
	h.mediaFile = source
	return h.Load(ctx)
}
