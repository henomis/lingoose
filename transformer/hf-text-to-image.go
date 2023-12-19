package transformer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	hfDefaultTextToImageModel = "stabilityai/stable-diffusion-xl-base-1.0"
)

type HFTextToImage struct {
	mediaFile  string
	token      string
	model      string
	httpClient *http.Client
}

type HFTextToImageRequest struct {
	Input string `json:"inputs"`
}

type HFTextToImageResponse struct {
	Output []byte
}

func NewHFTextToImage() *HFTextToImage {
	return &HFTextToImage{
		model:      hfDefaultTextToImageModel,
		token:      os.Getenv("HUGGING_FACE_HUB_TOKEN"),
		httpClient: http.DefaultClient,
	}
}

func (h *HFTextToImage) WithModel(model string) *HFTextToImage {
	h.model = model
	return h
}

func (h *HFTextToImage) WithToken(token string) *HFTextToImage {
	h.token = token
	return h
}

func (h *HFTextToImage) WithPersistImage(mediaFile string) *HFTextToImage {
	h.mediaFile = mediaFile
	return h
}

func (h *HFTextToImage) Transform(ctx context.Context, input string) (any, error) {
	respBody, err := h.hfTextToImageHTTPCall(ctx, input)
	if err != nil {
		return "", err
	}

	return respBody, nil
}

func (h *HFTextToImage) hfTextToImageHTTPCall(ctx context.Context, input string) ([]byte, error) {
	request := HFTextToImageRequest{
		Input: input,
	}

	buf, err := json.Marshal(request)
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error status code: %s", resp.Status)
	}

	if h.mediaFile != "" {
		f, errCreate := os.Create(h.mediaFile)
		if errCreate != nil {
			return nil, errCreate
		}
		defer f.Close()

		_, errCreate = io.Copy(f, resp.Body)
		if errCreate != nil {
			return nil, errCreate
		}

		return nil, nil
	}

	return io.ReadAll(resp.Body)
}
