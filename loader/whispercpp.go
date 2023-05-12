package loader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
)

type whisperCppLoader struct {
	loader loader

	ffmpegPath          string
	ffmpegArgs          []string
	whisperCppPath      string
	whisperCppArgs      []string
	whisperCppModelPath string
	filename            string
}

func NewWhisperCppLoader(filename string) *whisperCppLoader {
	return &whisperCppLoader{
		filename:            filename,
		ffmpegPath:          "/usr/bin/ffmpeg",
		ffmpegArgs:          []string{"-nostdin", "-f", "wav", "-ar", "16000", "-ac", "1", "-acodec", "pcm_s16le", "-"},
		whisperCppPath:      "./whisper.cpp/main",
		whisperCppArgs:      []string{},
		whisperCppModelPath: "./whisper.cpp/models/ggml-base.bin",
	}
}

func (t *whisperCppLoader) WithTextSplitter(textSplitter TextSplitter) *whisperCppLoader {
	t.loader.textSplitter = textSplitter
	return t
}

func (t *whisperCppLoader) WithFfmpegPath(ffmpegPath string) *whisperCppLoader {
	t.ffmpegPath = ffmpegPath
	return t
}

func (w *whisperCppLoader) WithWhisperCppPath(whisperCppPath string) *whisperCppLoader {
	w.whisperCppPath = whisperCppPath
	return w
}

func (w *whisperCppLoader) WithModel(whisperCppModelPath string) *whisperCppLoader {
	w.whisperCppModelPath = whisperCppModelPath
	return w
}

func (t *whisperCppLoader) WithArgs(whisperCppArgs []string) *whisperCppLoader {
	t.whisperCppArgs = whisperCppArgs
	return t
}

func (t *whisperCppLoader) Load(ctx context.Context) ([]document.Document, error) {

	err := t.validate()
	if err != nil {
		return nil, err
	}

	content, err := t.convertAndTrascribe(ctx)
	if err != nil {
		return nil, err
	}

	documents := []document.Document{
		{
			Content: content,
			Metadata: types.Meta{
				SourceMetadataKey: t.filename,
			},
		},
	}

	if t.loader.textSplitter != nil {
		documents = t.loader.textSplitter.SplitDocuments(documents)
	}

	return documents, nil
}

func (t *whisperCppLoader) validate() error {

	fileStat, err := os.Stat(t.filename)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	if fileStat.IsDir() {
		return fmt.Errorf("%s: %w", ErrorInternal, os.ErrNotExist)
	}

	return nil
}

func (whi *whisperCppLoader) convertAndTrascribe(ctx context.Context) (string, error) {

	ffmpegArgs := []string{"-i", whi.filename}
	ffmpegArgs = append(ffmpegArgs, whi.ffmpegArgs...)
	ffmpeg := exec.CommandContext(ctx, whi.ffmpegPath, ffmpegArgs...)

	whisperCppArgs := []string{"-m", whi.whisperCppModelPath, "-nt", "-f", "-"}
	whi.whisperCppArgs = append(whi.whisperCppArgs, whisperCppArgs...)
	whispercpp := exec.CommandContext(ctx, whi.whisperCppPath, whi.whisperCppArgs...)

	r, w := io.Pipe()
	ffmpeg.Stdout = w
	whispercpp.Stdin = r

	var out bytes.Buffer
	whispercpp.Stdout = &out

	err := ffmpeg.Start()
	if err != nil {
		return "", err
	}

	err = whispercpp.Start()
	if err != nil {
		return "", err
	}

	err = ffmpeg.Wait()
	if err != nil {
		return "", err
	}

	err = w.Close()
	if err != nil {
		return "", err
	}

	err = whispercpp.Wait()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}
