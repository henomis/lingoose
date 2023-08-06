package cohereembedder

import (
	"context"
	"github.com/cohere-ai/cohere-go"
	"github.com/henomis/lingoose/embedder"
	"net/http"
)

type Model string

const EmbedMultilingualV20 Model = "embed-multilingual-v2.0"
const EmbedEnglishLightV20 Model = "embed-english-light-v2.0"
const EmbedEnglishV20 Model = "embed-english-v2.0"

type CohereEmbedder struct {
	model  Model
	client *cohere.Client
}

// New Creates new Cohere Embedder Client
func New(model Model, apiKey string) *CohereEmbedder {

	client := &cohere.Client{
		APIKey:  apiKey,
		BaseURL: "https://api.cohere.ai/",
		Client:  *http.DefaultClient,
		Version: "2022-12-06",
	}

	return &CohereEmbedder{
		model:  model,
		client: client,
	}
}

// NewAndCheck Creates new Cohere Embedder Client and Validates the API Key
func NewAndCheck(model Model, apiKey string) (*CohereEmbedder, error) {
	client, err := cohere.CreateClient(apiKey)
	if err != nil {
		return nil, err
	}

	return &CohereEmbedder{
		model:  model,
		client: client,
	}, nil
}

func (c *CohereEmbedder) Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	response, err := c.client.Embed(cohere.EmbedOptions{
		Model:    string(c.model),
		Texts:    texts,
		Truncate: "",
	})
	if err != nil {
		return nil, err
	}

	var embeds []embedder.Embedding
	for _, embedding := range response.Embeddings {
		embeds = append(embeds, embedding)
	}
	return embeds, nil
}
