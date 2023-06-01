package loader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type hfImageToTextResponse struct {
	GeneratedText string `json:"generated_text"`
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
		return nil, fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	responseBytes, err := hfMediaHttpCall(ctx, h.token, h.model, h.mediaFile)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	responses := []*hfImageToTextResponse{}
	err = json.Unmarshal(responseBytes, &responses)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrorInternal, err)
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

func hfMediaHttpCall(ctx context.Context, token, model, mediaFile string) ([]byte, error) {
	buf, err := os.ReadFile(mediaFile)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hfAPIBaseURL+model, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.New("nil request created")
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = hfCheckHttpResponse(respBody)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func hfCheckHttpResponse(respJSON []byte) error {

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
