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

func (h *HuggingFace) textgenerationCompletion(ctx context.Context, prompts []string) ([]string, error) {

	numSequences := 1
	isTrue := true

	request := textGenerationRequest{
		Inputs: prompts,
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
		return nil, err
	}

	respBody, err := h.doRequest(ctx, jsonBuf, h.model)
	if err != nil {
		return nil, err
	}

	tgrespsRaw := make([][]*textGenerationResponseSequence, len(request.Inputs))
	err = json.Unmarshal(respBody, &tgrespsRaw)
	if err != nil {
		return nil, err
	}
	if len(tgrespsRaw) != len(request.Inputs) {
		return nil, fmt.Errorf("%s: expected %d responses, got %d; response=%s", ErrHuggingFaceCompletion, len(request.Inputs), len(tgrespsRaw), string(respBody))
	}

	outputs := make([]string, len(request.Inputs))
	for i := range tgrespsRaw {
		for _, t := range tgrespsRaw[i] {
			output := strings.TrimLeft(t.GeneratedText, prompts[i])
			output = strings.TrimSpace(output)
			outputs[i] = output
			if h.verbose {
				debugCompletion(prompts[i], outputs[i])
			}
		}
	}

	return outputs, nil
}
