package loader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
)

const (
	hfAPIBaseURL              = "https://api-inference.huggingface.co/models/"
	hfDefaultImageToTextModel = "nlpconnect/vit-gpt2-image-captioning"
)

type hfImageToText struct {
	loader loader

	mediaFile string
	token     string
	model     string
}

func NewHFImageToTextLoader(mediaFile string) *hfImageToText {
	return &hfImageToText{
		mediaFile: mediaFile,
		model:     hfDefaultImageToTextModel,
		token:     os.Getenv("HUGGING_FACE_HUB_TOKEN"),
	}
}

func (h *hfImageToText) WithToken(token string) *hfImageToText {
	h.token = token
	return h
}

func (h *hfImageToText) WithModel(model string) *hfImageToText {
	h.model = model
	return h
}

func (h *hfImageToText) WithTextSplitter(textSplitter TextSplitter) *hfImageToText {
	h.loader.textSplitter = textSplitter
	return h
}

func (h *hfImageToText) Load(ctx context.Context) ([]document.Document, error) {

	err := isFile(h.mediaFile)
	if err != nil {
		return nil, err
	}

	responses, err := h.httpCall(ctx)
	if err != nil {
		return nil, err
	}

	var documents []document.Document
	document := document.Document{
		Content: "",
		Metadata: types.Meta{
			"source": h.mediaFile,
		},
	}
	for _, r := range responses {
		if r.GeneratedText != "" {
			document.Content += r.GeneratedText + "\n"
		}
	}

	document.Content = strings.TrimSpace(document.Content)
	documents = append(documents, document)

	if h.loader.textSplitter != nil {
		documents = h.loader.textSplitter.SplitDocuments(documents)
	}

	return documents, nil
}

type ImageToTextResponse struct {
	GeneratedText string `json:"generated_text"`
}

func (h *hfImageToText) httpCall(ctx context.Context) ([]*ImageToTextResponse, error) {
	buf, err := os.ReadFile(h.mediaFile)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hfAPIBaseURL+h.model, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.New("nil request created")
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer "+h.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = checkResponse(respBody)
	if err != nil {
		return nil, err
	}

	resps := []*ImageToTextResponse{}
	err = json.Unmarshal(respBody, &resps)
	if err != nil {
		return nil, err
	}

	return resps, nil
}

func checkResponse(respJSON []byte) error {

	type apiError struct {
		Error string `json:"error,omitempty"`
	}

	type apiErrors struct {
		Errors []string `json:"error,omitempty"`
	}

	{
		buf := make([]byte, len(respJSON))
		copy(buf, respJSON)
		apiErr := apiError{}
		err := json.Unmarshal(buf, &apiErr)
		if err != nil {
			return nil
		}
		if apiErr.Error != "" {
			return errors.New(string(respJSON))
		}
	}

	// Check for multiple errors
	{
		buf := make([]byte, len(respJSON))
		copy(buf, respJSON)
		apiErrs := apiErrors{}
		err := json.Unmarshal(buf, &apiErrs)
		if err != nil {
			return nil
		}
		if apiErrs.Errors != nil {
			return errors.New(string(respJSON))
		}
	}

	return nil
}
