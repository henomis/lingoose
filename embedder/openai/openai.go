package openaiembedder

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/henomis/lingoose/embedder"
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
}

func New(model Model) (*openAIEmbedder, error) {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	return &openAIEmbedder{
		openAIClient: openai.NewClient(openAIKey),
		model:        model,
	}, nil
}

func (o *openAIEmbedder) Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	maxTokens := o.getMaxTokens()

	var embeddings []embedder.Embedding
	for _, text := range texts {
		embedding, err := o.safeEmbed(ctx, text, maxTokens)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", embedder.ErrCreateEmbedding, err)
		}

		embeddings = append(embeddings, embedding)
	}

	return embeddings, nil
}

func (o *openAIEmbedder) safeEmbed(ctx context.Context, text string, maxTokens int) (embedder.Embedding, error) {

	sanitizedText := text
	if strings.HasSuffix(o.model.String(), "001") {
		sanitizedText = strings.ReplaceAll(text, "\n", " ")
	}

	chunkedText, err := o.chunkText(sanitizedText, maxTokens)
	if err != nil {
		return nil, err
	}

	embeddingsForChunks, chunkLens, err := o.getEmebeddingsForChunks(ctx, chunkedText)
	if err != nil {
		return nil, err
	}

	return normalizeEmbeddings(embeddingsForChunks, chunkLens), nil

}

func (o *openAIEmbedder) chunkText(text string, maxTokens int) ([]string, error) {

	tokens, err := o.textToTokens(text)
	if err != nil {
		return nil, err
	}

	var textChunks []string
	for i := 0; i < len(tokens); i += maxTokens {
		end := i + maxTokens
		if end > len(tokens) {
			end = len(tokens)
		}

		textChunk, err := o.tokensToText(tokens[i:end])
		if err != nil {
			return nil, err
		}

		textChunks = append(textChunks, textChunk)
	}

	return textChunks, nil
}

func (o *openAIEmbedder) getEmebeddingsForChunks(ctx context.Context, chunks []string) ([]embedder.Embedding, []float64, error) {

	chunkLens := []float64{}

	embeddingsForChunks, err := o.openAICreateEmebeddings(ctx, chunks)
	if err != nil {
		return nil, nil, err
	}

	for _, chunk := range chunks {
		chunkLens = append(chunkLens, float64(len(chunk)))
	}

	return embeddingsForChunks, chunkLens, nil

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
		return nil, err
	}

	var embeddings []embedder.Embedding

	for _, obj := range resp.Data {
		embeddings = append(embeddings, float32ToFloat64(obj.Embedding))
	}

	return embeddings, nil
}

func float32ToFloat64(slice []float32) []float64 {
	newSlice := make([]float64, len(slice))
	for i, v := range slice {
		newSlice[i] = float64(v)
	}
	return newSlice
}
