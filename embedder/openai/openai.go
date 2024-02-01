package openaiembedder

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/henomis/lingoose/embedder"
	"github.com/sashabaranov/go-openai"
)

type Model = openai.EmbeddingModel

const (
	AdaEmbeddingV2  Model = openai.AdaEmbeddingV2
	SmallEmbedding3 Model = openai.SmallEmbedding3
	LargeEmbedding3 Model = openai.LargeEmbedding3
)

type OpenAIEmbedder struct {
	openAIClient *openai.Client
	model        Model
}

func New(model Model) *OpenAIEmbedder {
	openAIKey := os.Getenv("OPENAI_API_KEY")

	return &OpenAIEmbedder{
		openAIClient: openai.NewClient(openAIKey),
		model:        model,
	}
}

// WithClient sets the OpenAI client to use for the embedder
func (o *OpenAIEmbedder) WithClient(client *openai.Client) *OpenAIEmbedder {
	o.openAIClient = client
	return o
}

// Embed returns the embeddings for the given texts
func (o *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	maxTokens := o.getMaxTokens()

	embeddings, err := o.concurrentEmbed(ctx, texts, maxTokens)
	if err != nil {
		return nil, err
	}

	return embeddings, nil
}

func (o *OpenAIEmbedder) concurrentEmbed(
	ctx context.Context,
	texts []string,
	maxTokens int,
) ([]embedder.Embedding, error) {
	type indexedEmbeddings struct {
		index     int
		embedding embedder.Embedding
		err       error
	}

	var embeddings []indexedEmbeddings
	embeddingsChan := make(chan indexedEmbeddings, len(texts))

	for i, text := range texts {
		go func(ctx context.Context, i int, text string, maxTokens int) {
			embedding, err := o.safeEmbed(ctx, text, maxTokens)

			embeddingsChan <- indexedEmbeddings{
				index:     i,
				embedding: embedding,
				err:       err,
			}
		}(ctx, i, text, maxTokens)
	}

	var err error
	for i := 0; i < len(texts); i++ {
		embedding := <-embeddingsChan
		if embedding.err != nil {
			err = embedding.err
			continue
		}
		embeddings = append(embeddings, embedding)
	}

	if err != nil {
		return nil, err
	}

	sort.Slice(embeddings, func(i, j int) bool {
		return embeddings[i].index < embeddings[j].index
	})

	var result []embedder.Embedding
	for _, embedding := range embeddings {
		result = append(result, embedding.embedding)
	}

	return result, nil
}

func (o *OpenAIEmbedder) safeEmbed(ctx context.Context, text string, maxTokens int) (embedder.Embedding, error) {
	sanitizedText := text
	if strings.HasSuffix(string(o.model), "001") {
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

func (o *OpenAIEmbedder) chunkText(text string, maxTokens int) ([]string, error) {
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

		textChunk, errToken := o.tokensToText(tokens[i:end])
		if errToken != nil {
			return nil, errToken
		}

		textChunks = append(textChunks, textChunk)
	}

	return textChunks, nil
}

func (o *OpenAIEmbedder) getEmebeddingsForChunks(
	ctx context.Context,
	chunks []string,
) ([]embedder.Embedding, []float64, error) {
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

func (o *OpenAIEmbedder) openAICreateEmebeddings(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	resp, err := o.openAIClient.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Input: texts,
			Model: o.model,
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
