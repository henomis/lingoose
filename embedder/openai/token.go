package openaiembedder

import (
	"github.com/pkoukk/tiktoken-go"
)

func (o *OpenAIEmbedder) textToTokens(text string) ([]int, error) {
	tokenizer, err := tiktoken.EncodingForModel(o.model.String())
	if err != nil {
		return nil, err
	}

	return tokenizer.Encode(text, nil, nil), nil
}

func (o *OpenAIEmbedder) getMaxTokens() int {

	if tiktoken.MODEL_TO_ENCODING[o.model.String()] == "cl100k_base" {
		return 8191
	}

	return 2046
}

func (o *OpenAIEmbedder) tokensToText(tokens []int) (string, error) {
	tokenizer, err := tiktoken.EncodingForModel(o.model.String())
	if err != nil {
		return "", err
	}

	return tokenizer.Decode(tokens), nil
}
