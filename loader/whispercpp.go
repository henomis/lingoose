package loader

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"regexp"

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

var whisperSanitizeRegexp = regexp.MustCompile(`\[.*?\]`)

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

func (w *whisperCppLoader) WithTextSplitter(textSplitter TextSplitter) *whisperCppLoader {
	w.loader.textSplitter = textSplitter
	return w
}

func (w *whisperCppLoader) WithFfmpegPath(ffmpegPath string) *whisperCppLoader {
	w.ffmpegPath = ffmpegPath
	return w
}

func (w *whisperCppLoader) WithWhisperCppPath(whisperCppPath string) *whisperCppLoader {
	w.whisperCppPath = whisperCppPath
	return w
}

func (w *whisperCppLoader) WithModel(whisperCppModelPath string) *whisperCppLoader {
	w.whisperCppModelPath = whisperCppModelPath
	return w
}

func (w *whisperCppLoader) WithArgs(whisperCppArgs []string) *whisperCppLoader {
	w.whisperCppArgs = whisperCppArgs
	return w
}

func (w *whisperCppLoader) Load(ctx context.Context) ([]document.Document, error) {

	err := isFile(w.ffmpegPath)
	if err != nil {
		return nil, err
	}

	err = isFile(w.whisperCppPath)
	if err != nil {
		return nil, err
	}

	err = isFile(w.filename)
	if err != nil {
		return nil, err
	}

	content, err := w.convertAndTrascribe(ctx)
	if err != nil {
		return nil, err
	}

	documents := []document.Document{
		{
			Content: content,
			Metadata: types.Meta{
				SourceMetadataKey: w.filename,
			},
		},
	}

	if w.loader.textSplitter != nil {
		documents = w.loader.textSplitter.SplitDocuments(documents)
	}

	return documents, nil
}

func (w *whisperCppLoader) convertAndTrascribe(ctx context.Context) (string, error) {

	ffmpegArgs := []string{"-i", w.filename}
	ffmpegArgs = append(ffmpegArgs, w.ffmpegArgs...)
	ffmpeg := exec.CommandContext(ctx, w.ffmpegPath, ffmpegArgs...)

	whisperCppArgs := []string{"-m", w.whisperCppModelPath, "-nt", "-f", "-"}
	whisperCppArgs = append(w.whisperCppArgs, whisperCppArgs...)

	whispercpp := exec.CommandContext(ctx, w.whisperCppPath, whisperCppArgs...)

	pipeReader, pipeWriter := io.Pipe()
	ffmpeg.Stdout = pipeWriter
	whispercpp.Stdin = pipeReader

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

	err = pipeWriter.Close()
	if err != nil {
		return "", err
	}

	err = whispercpp.Wait()
	if err != nil {
		return "", err
	}

	return whisperSanitizeRegexp.ReplaceAllString(out.String(), ""), nil
}
