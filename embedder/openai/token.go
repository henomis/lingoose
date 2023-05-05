package openaiembedder

import (
	"github.com/pkoukk/tiktoken-go"
)

func (o *openAIEmbedder) textToTokens(text string) ([]int, error) {
	return o.tiktoken.Encode(text, nil, nil), nil
}

func (o *openAIEmbedder) getMaxTokens() int {

	if tiktoken.MODEL_TO_ENCODING[o.model.String()] == "cl100k_base" {
		return 8191
	}

	return 2046
}

func (o *openAIEmbedder) tokensToText(tokens []int) (string, error) {
	return o.tiktoken.Decode(tokens), nil
}
