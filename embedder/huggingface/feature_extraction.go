package huggingfaceembedder

import (
	"context"
	"encoding/json"

	"github.com/henomis/lingoose/embedder"
)

type options struct {
	WaitForModel *bool `json:"wait_for_model,omitempty"`
}

type featureExtractionRequest struct {
	Inputs  []string `json:"inputs,omitempty"`
	Options options  `json:"options,omitempty"`
}

func (h *huggingFaceEmbedder) featureExtraction(ctx context.Context, text []string) ([]embedder.Embedding, error) {

	isTrue := true
	request := featureExtractionRequest{
		Inputs: text,
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

	var embeddings []embedder.Embedding
	err = json.Unmarshal(respBody, &embeddings)
	if err != nil {
		return nil, err
	}

	return embeddings, nil
}
