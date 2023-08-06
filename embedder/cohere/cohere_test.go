package cohereembedder

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func Test_CohereEmbedder(t *testing.T) {

	apiKey := os.Getenv("COHERE_API_KEY")

	ce := New(EmbedEnglishLightV20, apiKey)

	texts := []string{
		"Hello, LinGoose",
		"مرحباً بالعالم!",
	}

	result, err := ce.Embed(context.Background(), texts)
	if err != nil {
		t.Errorf("CohereEmbedder.Embed() error = %v", err)
		return
	}

	if len(result) != len(texts) {
		t.Errorf("CohereEmbedder.Embed() error = %v", fmt.Errorf("got %v embeddings instead of %v", len(result), len(texts)))
		return
	}
}
