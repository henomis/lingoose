package huggingface

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

const APIBaseURL = "https://api-inference.huggingface.co/models/"

const (
	ErrHuggingFaceCompletion = "huggingface completion error"
)

type HuggingFaceMode int

const (
	HuggingFaceModeCoversational HuggingFaceMode = iota
)

type huggingFace struct {
	mode        HuggingFaceMode
	token       string
	model       string
	temperature float32
	maxLength   *int
	minLength   *int
	topK        *int
	topP        *float32
	verbose     bool
}

type Options struct {
	UseGPU       *bool `json:"use_gpu,omitempty"`
	UseCache     *bool `json:"use_cache,omitempty"`
	WaitForModel *bool `json:"wait_for_model,omitempty"`
}

func New(model string, temperature float32, verbose bool) *huggingFace {
	return &huggingFace{
		mode:        HuggingFaceModeCoversational,
		token:       os.Getenv("HUGGING_FACE_HUB_TOKEN"),
		model:       model,
		temperature: temperature,
		verbose:     verbose,
	}
}

func (h *huggingFace) WithModel(model string) *huggingFace {
	h.model = model
	return h
}

func (h *huggingFace) WithTemperature(temperature float32) *huggingFace {
	h.temperature = temperature
	return h
}

func (h *huggingFace) WithMaxLength(maxLength int) *huggingFace {
	h.maxLength = &maxLength
	return h
}

func (h *huggingFace) WithMinLength(minLength int) *huggingFace {
	h.minLength = &minLength
	return h
}

func (h *huggingFace) WithToken(token string) *huggingFace {
	h.token = token
	return h
}

func (h *huggingFace) WithVerbose(verbose bool) *huggingFace {
	h.verbose = verbose
	return h
}

func (h *huggingFace) WithTopK(topK int) *huggingFace {
	h.topK = &topK
	return h
}

func (h *huggingFace) WithTopP(topP float32) *huggingFace {
	h.topP = &topP
	return h
}

func (h *huggingFace) WithMode(mode HuggingFaceMode) *huggingFace {
	h.mode = mode
	return h
}

func (h *huggingFace) Completion(ctx context.Context, prompt string) (string, error) {

	var output string
	var err error
	switch h.mode {
	case HuggingFaceModeCoversational:
		fallthrough
	default:
		output, err = h.conversationalCompletion(ctx, prompt)
	}

	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrHuggingFaceCompletion, err)
	}

	return output, nil
}

func (h *huggingFace) doRequest(ctx context.Context, jsonBody []byte, model string) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, APIBaseURL+model, bytes.NewBuffer(jsonBody))
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = checkRespForError(respBody)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

type apiError struct {
	Error string `json:"error,omitempty"`
}

type apiErrors struct {
	Errors []string `json:"error,omitempty"`
}

func checkRespForError(respJSON []byte) error {
	{
		buf := make([]byte, len(respJSON))
		copy(buf, respJSON)
		apiErr := apiError{}
		err := json.Unmarshal(buf, &apiErr)
		if err != nil {
			return err
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
			return err
		}
		if apiErrs.Errors != nil {
			return errors.New(string(respJSON))
		}
	}

	return nil
}

func debugCompletion(prompt string, content string) {
	fmt.Printf("---USER---\n%s\n", prompt)
	fmt.Printf("---AI---\n%s\n", content)
}
