package llamacpp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/henomis/lingoose/chat"
)

const (
	DefaultLlamaCppMaxTokens   = 256
	DefaultLlamaCppTemperature = 0.8
)

type Llamacpp struct {
	llamacppPath string
	llamacppArgs []string
	modelPath    string
	temperature  float32
	maxTokens    int
	verbose      bool
}

var llamacppSanitizeRegexp = regexp.MustCompile(`\[.*?\]`)

func NewCompletion() *Llamacpp {
	return &Llamacpp{
		llamacppPath: "./llama.cpp/main",
		modelPath:    "./llama.cpp/models/7B/ggml-model-q4_0.bin",
		llamacppArgs: []string{},
		temperature:  DefaultLlamaCppTemperature,
		maxTokens:    DefaultLlamaCppMaxTokens,
	}
}

func (l *Llamacpp) WithModel(modelPath string) *Llamacpp {
	l.modelPath = modelPath
	return l
}

func (l *Llamacpp) WithTemperature(temperature float32) *Llamacpp {
	l.temperature = temperature
	return l
}

func (l *Llamacpp) WithMaxTokens(maxTokens int) *Llamacpp {
	l.maxTokens = maxTokens
	return l
}

func (l *Llamacpp) WithVerbose(verbose bool) *Llamacpp {
	l.verbose = verbose
	return l
}

func (l *Llamacpp) WithLlamaCppPath(llamacppPath string) *Llamacpp {
	l.llamacppPath = llamacppPath
	return l
}

func (l *Llamacpp) WithArgs(llamacppArgs []string) *Llamacpp {
	l.llamacppArgs = llamacppArgs
	return l
}

func (l *Llamacpp) Completion(ctx context.Context, prompt string) (string, error) {

	_, err := os.Stat(l.llamacppPath)
	if err != nil {
		return "", err
	}

	llamacppArgs := []string{
		"-m", l.modelPath,
		"-p", prompt,
		"-n", fmt.Sprintf("%d", l.maxTokens),
		"--temp", fmt.Sprintf("%.2f", l.temperature),
	}
	llamacppArgs = append(llamacppArgs, l.llamacppArgs...)

	//nolint:gosec
	out, err := exec.CommandContext(ctx, l.llamacppPath, llamacppArgs...).Output()
	if err != nil {
		return "", err
	}

	if l.verbose {
		fmt.Printf("---USER---\n%s\n", prompt)
		fmt.Printf("---AI---\n%s\n", out)
	}

	return llamacppSanitizeRegexp.ReplaceAllString(string(out), ""), nil
}

func (l *Llamacpp) Chat(ctx context.Context, prompt *chat.Chat) (string, error) {
	_ = ctx
	_ = prompt
	return "", fmt.Errorf("not implemented")
}
