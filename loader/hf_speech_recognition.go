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

type hfSpeechRecognition struct {
	loader loader

	mediaFile string
	token     string
	model     string
}

type hfSpeechRecognitionResponse struct {
	Text string `json:"text,omitempty"`
}

func NewHFSpeechRecognitionLoader(mediaFile string) *hfSpeechRecognition {
	return &hfSpeechRecognition{
		mediaFile: mediaFile,
		model:     hfDefaultSpeechRecognitionModel,
		token:     os.Getenv("HUGGING_FACE_HUB_TOKEN"),
	}
}

func (h *hfSpeechRecognition) WithToken(token string) *hfSpeechRecognition {
	h.token = token
	return h
}

func (h *hfSpeechRecognition) WithModel(model string) *hfSpeechRecognition {
	h.model = model
	return h
}

func (h *hfSpeechRecognition) WithTextSplitter(textSplitter TextSplitter) *hfSpeechRecognition {
	h.loader.textSplitter = textSplitter
	return h
}

func (h *hfSpeechRecognition) Load(ctx context.Context) ([]document.Document, error) {

	err := isFile(h.mediaFile)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	responseBytes, err := hfMediaHttpCall(ctx, h.token, h.model, h.mediaFile)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	response := hfSpeechRecognitionResponse{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrorInternal, err)
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
