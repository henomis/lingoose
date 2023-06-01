package huggingfaceembedder

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const APIBaseURL = "https://api-inference.huggingface.co/pipeline/feature-extraction/"

func (h *huggingFaceEmbedder) doRequest(ctx context.Context, jsonBody []byte, model string) ([]byte, error) {

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
