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

type llamacpp struct {
	llamacppPath string
	llamacppArgs []string
	modelPath    string
	temperature  float32
	maxTokens    int
	verbose      bool
}

var llamacppSanitizeRegexp = regexp.MustCompile(`\[.*?\]`)

func NewCompletion() *llamacpp {
	return &llamacpp{
		llamacppPath: "./llama.cpp/main",
		modelPath:    "./llama.cpp/models/7B/ggml-model-q4_0.bin",
		llamacppArgs: []string{},
		temperature:  DefaultLlamaCppTemperature,
		maxTokens:    DefaultLlamaCppMaxTokens,
	}
}

func (l *llamacpp) WithModel(modelPath string) *llamacpp {
	l.modelPath = modelPath
	return l
}

func (l *llamacpp) WithTemperature(temperature float32) *llamacpp {
	l.temperature = temperature
	return l
}

func (l *llamacpp) WithMaxTokens(maxTokens int) *llamacpp {
	l.maxTokens = maxTokens
	return l
}

func (l *llamacpp) WithVerbose(verbose bool) *llamacpp {
	l.verbose = verbose
	return l
}

func (l *llamacpp) WithLlamaCppPath(llamacppPath string) *llamacpp {
	l.llamacppPath = llamacppPath
	return l
}

func (l *llamacpp) WithArgs(llamacppArgs []string) *llamacpp {
	l.llamacppArgs = llamacppArgs
	return l
}

func (l *llamacpp) Completion(ctx context.Context, prompt string) (string, error) {

	_, err := os.Stat(l.llamacppPath)
	if err != nil {
		return "", err
	}

	llamacppArgs := []string{"-m", l.modelPath, "-p", prompt, "-n", fmt.Sprintf("%d", l.maxTokens), "--temp", fmt.Sprintf("%.2f", l.temperature)}
	llamacppArgs = append(llamacppArgs, l.llamacppArgs...)

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

func (l *llamacpp) Chat(ctx context.Context, prompt *chat.Chat) (string, error) {
	return "", fmt.Errorf("not implemented")
}
