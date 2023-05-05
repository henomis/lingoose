package openaiembedder

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/henomis/lingoose/embedder"
	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
)

type Model int

const (
	Unknown Model = iota
	AdaSimilarity
	BabbageSimilarity
	CurieSimilarity
	DavinciSimilarity
	AdaSearchDocument
	AdaSearchQuery
	BabbageSearchDocument
	BabbageSearchQuery
	CurieSearchDocument
	CurieSearchQuery
	DavinciSearchDocument
	DavinciSearchQuery
	AdaCodeSearchCode
	AdaCodeSearchText
	BabbageCodeSearchCode
	BabbageCodeSearchText
	AdaEmbeddingV2
)

func (m Model) String() string {
	return modelToString[m]
}

var modelToString = map[Model]string{
	AdaSimilarity:         "text-similarity-ada-001",
	BabbageSimilarity:     "text-similarity-babbage-001",
	CurieSimilarity:       "text-similarity-curie-001",
	DavinciSimilarity:     "text-similarity-davinci-001",
	AdaSearchDocument:     "text-search-ada-doc-001",
	AdaSearchQuery:        "text-search-ada-query-001",
	BabbageSearchDocument: "text-search-babbage-doc-001",
	BabbageSearchQuery:    "text-search-babbage-query-001",
	CurieSearchDocument:   "text-search-curie-doc-001",
	CurieSearchQuery:      "text-search-curie-query-001",
	DavinciSearchDocument: "text-search-davinci-doc-001",
	DavinciSearchQuery:    "text-search-davinci-query-001",
	AdaCodeSearchCode:     "code-search-ada-code-001",
	AdaCodeSearchText:     "code-search-ada-text-001",
	BabbageCodeSearchCode: "code-search-babbage-code-001",
	BabbageCodeSearchText: "code-search-babbage-text-001",
	AdaEmbeddingV2:        "text-embedding-ada-002",
}

type openAIEmbedder struct {
	openAIClient *openai.Client
	model        Model
	tiktoken     *tiktoken.Tiktoken
}

func New(model Model) (*openAIEmbedder, error) {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	tiktoken, err := tiktoken.EncodingForModel(model.String())
	if err != nil {
		return nil, err
	}

	return &openAIEmbedder{
		openAIClient: openai.NewClient(openAIKey),
		model:        model,
		tiktoken:     tiktoken,
	}, nil
}

func (t *openAIEmbedder) openAICreateEmebeddings(ctx context.Context, texts []string) ([]embedder.Embedding, error) {

	resp, err := t.openAIClient.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Input: texts,
			Model: openai.EmbeddingModel(t.model),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", embedder.ErrCreateEmbedding, err)
	}

	var embeddings []embedder.Embedding

	for _, obj := range resp.Data {
		embeddings = append(embeddings, float32ToFloat64(obj.Embedding))
	}

	return embeddings, nil
}

func (o *openAIEmbedder) Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	return o.safeEmbed(ctx, texts, o.getMaxTokens())
}

func (o *openAIEmbedder) safeEmbed(ctx context.Context, texts []string, maxTokens int) ([]embedder.Embedding, error) {

	var embeddings []embedder.Embedding
	for _, text := range texts {

		formattedText := text
		if strings.HasSuffix(o.model.String(), "001") {
			formattedText = strings.ReplaceAll(text, "\n", " ")
		}

		chunkedText, err := o.splitText(formattedText, maxTokens)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", embedder.ErrCreateEmbedding, err)
		}

		chunkEmbeddings, chunkLens, err := o.getEmebeddingsForChunks(ctx, chunkedText)
		if err != nil {
			return nil, err
		}

		embeddings = append(embeddings, normalizeEmbeddings(chunkEmbeddings, chunkLens))

	}

	return embeddings, nil
}

func (o *openAIEmbedder) splitText(text string, maxTokens int) ([]string, error) {

	tokens, err := o.textToTokens(text)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", embedder.ErrCreateEmbedding, err)
	}

	var chunkedText []string
	for i := 0; i < len(tokens); i += maxTokens {
		end := i + maxTokens
		if end > len(tokens) {
			end = len(tokens)
		}

		text, err := o.tokensToText(tokens[i:end])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", embedder.ErrCreateEmbedding, err)
		}

		chunkedText = append(chunkedText, text)
	}

	return chunkedText, nil
}

func (o *openAIEmbedder) getEmebeddingsForChunks(ctx context.Context, chunks []string) ([]embedder.Embedding, []float64, error) {

	chunkLens := []float64{}

	chunkEmbeddings, err := o.openAICreateEmebeddings(ctx, chunks)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", embedder.ErrCreateEmbedding, err)
	}

	for _, chunk := range chunks {
		chunkLens = append(chunkLens, float64(len(chunk)))
	}

	return chunkEmbeddings, chunkLens, nil

}

func float32ToFloat64(slice []float32) []float64 {
	newSlice := make([]float64, len(slice))
	for i, v := range slice {
		newSlice[i] = float64(v)
	}
	return newSlice
}
