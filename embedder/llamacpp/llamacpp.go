package llamacppembedder

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"

	"github.com/henomis/lingoose/embedder"
)

type LlamaCppEmbedder struct {
	llamacppPath string
	llamacppArgs []string
	modelPath    string
}

type output struct {
	Object string `json:"object"`
	Data   []data `json:"data"`
}
type data struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

func New() *LlamaCppEmbedder {
	return &LlamaCppEmbedder{
		llamacppPath: "./llama.cpp/embedding",
		modelPath:    "./llama.cpp/models/7B/ggml-model-q4_0.bin",
		llamacppArgs: []string{},
	}
}

// WithLlamaCppPath sets the path to the llamacpp binary
func (l *LlamaCppEmbedder) WithLlamaCppPath(llamacppPath string) *LlamaCppEmbedder {
	l.llamacppPath = llamacppPath
	return l
}

// WithModel sets the model to use for the embedder
func (l *LlamaCppEmbedder) WithModel(modelPath string) *LlamaCppEmbedder {
	l.modelPath = modelPath
	return l
}

// WithArgs sets the args to pass to the llamacpp binary
func (l *LlamaCppEmbedder) WithArgs(llamacppArgs []string) *LlamaCppEmbedder {
	l.llamacppArgs = llamacppArgs
	return l
}

// Embed returns the embeddings for the given texts
func (l *LlamaCppEmbedder) Embed(ctx context.Context, texts []string) ([]embedder.Embedding, error) {
	embeddings := make([]embedder.Embedding, len(texts))
	for i, text := range texts {
		embedding, err := l.embed(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}

func (l *LlamaCppEmbedder) embed(ctx context.Context, text string) (embedder.Embedding, error) {
	_, err := os.Stat(l.llamacppPath)
	if err != nil {
		return nil, err
	}

	llamacppArgs := []string{"-m", l.modelPath, "--embd-output-format", "json", "-p", text}
	llamacppArgs = append(llamacppArgs, l.llamacppArgs...)

	//nolint:gosec
	out, err := exec.CommandContext(ctx, l.llamacppPath, llamacppArgs...).Output()
	if err != nil {
		return nil, err
	}

	return parseEmbeddings(string(out))
}

func parseEmbeddings(str string) (embedder.Embedding, error) {
	var output output
	err := json.Unmarshal([]byte(str), &output)
	if err != nil {
		return nil, err
	}

	if len(output.Data) != 1 {
		return nil, errors.New("no embeddings found")
	}

	return output.Data[0].Embedding, nil
}
