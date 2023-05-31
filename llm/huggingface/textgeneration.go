package huggingface

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type textGenerationRequest struct {
	Inputs     []string                 `json:"inputs,omitempty"`
	Parameters textGenerationParameters `json:"parameters,omitempty"`
	Options    options                  `json:"options,omitempty"`
}

type textGenerationParameters struct {
	TopK               *int     `json:"top_k,omitempty"`
	TopP               *float64 `json:"top_p,omitempty"`
	Temperature        *float32 `json:"temperature,omitempty"`
	RepetitionPenalty  *float64 `json:"repetition_penalty,omitempty"`
	MaxNewTokens       *int     `json:"max_new_tokens,omitempty"`
	MaxTime            *float64 `json:"max_time,omitempty"`
	ReturnFullText     *bool    `json:"return_full_text,omitempty"`
	NumReturnSequences *int     `json:"num_return_sequences,omitempty"`
}

type textGenerationResponseSequence struct {
	GeneratedText string `json:"generated_text,omitempty"`
}

func (tgs textGenerationResponseSequence) String() string {
	return tgs.GeneratedText
}

func (h *huggingFace) textgenerationCompletion(ctx context.Context, prompt string) (string, error) {

	numSequences := 1
	isTrue := true

	request := textGenerationRequest{
		Inputs: []string{prompt},
		Parameters: textGenerationParameters{
			Temperature:        &h.temperature,
			TopK:               h.topK,
			NumReturnSequences: &numSequences,
		},
		Options: options{
			WaitForModel: &isTrue,
		},
	}

	jsonBuf, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	respBody, err := h.doRequest(ctx, jsonBuf, h.model)
	if err != nil {
		return "", err
	}

	tgrespsRaw := make([][]*textGenerationResponseSequence, len(request.Inputs))
	err = json.Unmarshal(respBody, &tgrespsRaw)
	if err != nil {
		return "", err
	}
	if len(tgrespsRaw) != len(request.Inputs) {
		return "", fmt.Errorf("%s: expected %d responses, got %d; response=%s", ErrHuggingFaceCompletion, len(request.Inputs), len(tgrespsRaw), string(respBody))
	}

	output := ""
	for i := range tgrespsRaw {
		for _, t := range tgrespsRaw[i] {
			output += t.GeneratedText
		}
	}

	output = strings.TrimLeft(output, prompt)
	output = strings.TrimSpace(output)

	debugCompletion(prompt, output)

	return output, nil
}
