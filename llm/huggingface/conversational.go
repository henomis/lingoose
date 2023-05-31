package huggingface

import (
	"context"
	"encoding/json"
)

type conversationalRequest struct {
	Inputs     conversationalInputs     `json:"inputs,omitempty"`
	Parameters conversationalParameters `json:"parameters,omitempty"`
	Options    options                  `json:"options,omitempty"`
}

type conversationalParameters struct {
	MinLength         *int     `json:"min_length,omitempty"`
	MaxLength         *int     `json:"max_length,omitempty"`
	TopK              *int     `json:"top_k,omitempty"`
	TopP              *float32 `json:"top_p,omitempty"`
	Temperature       *float32 `json:"temperature,omitempty"`
	RepetitionPenalty *float32 `json:"repetitionpenalty,omitempty"`
	MaxTime           *float32 `json:"maxtime,omitempty"`
}

type conversationalInputs struct {
	Text               string   `json:"text,omitempty"`
	GeneratedResponses []string `json:"generated_responses,omitempty"`
	PastUserInputs     []string `json:"past_user_inputs,omitempty"`
}

type conversationalResponse struct {
	GeneratedText string       `json:"generated_text,omitempty"`
	Conversation  conversation `json:"conversation,omitempty"`
}

type conversation struct {
	GeneratedResponses []string `json:"generated_responses,omitempty"`
	PastUserInputs     []string `json:"past_user_inputs,omitempty"`
}

func (h *huggingFace) conversationalCompletion(ctx context.Context, prompt string) (string, error) {

	isTrue := true
	request := conversationalRequest{
		Inputs: conversationalInputs{
			Text: prompt,
		},
		Parameters: conversationalParameters{
			Temperature: &h.temperature,
			MinLength:   h.minLength,
			MaxLength:   h.maxLength,
			TopK:        h.topK,
			TopP:        h.topP,
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

	cresp := conversationalResponse{}
	err = json.Unmarshal(respBody, &cresp)
	if err != nil {
		return "", err
	}

	debugCompletion(prompt, cresp.GeneratedText)

	return cresp.GeneratedText, nil
}
