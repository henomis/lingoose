package huggingface

import (
	"context"
	"fmt"
	"os"
)

const APIBaseURL = "https://api-inference.huggingface.co/models/"

const (
	ErrHuggingFaceCompletion = "huggingface completion error"
)

type HuggingFaceMode int

const (
	HuggingFaceModeCoversational HuggingFaceMode = iota
	HuggingFaceModeTextGeneration
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
	case HuggingFaceModeTextGeneration:
		output, err = h.textgenerationCompletion(ctx, prompt)
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
