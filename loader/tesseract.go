package loader

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
)

var (
	ErrTesseractNotFound = fmt.Errorf("pdftotext not found")
	defaultTesseractPath = "/usr/bin/tesseract"
)

type tesseractLoader struct {
	loader loader

	tesseractPath string
	tesseractArgs []string
	filename      string
}

func NewTesseractLoader(filename string) *tesseractLoader {
	return &tesseractLoader{
		tesseractPath: defaultTesseractPath,
		tesseractArgs: []string{},
		filename:      filename,
	}
}

func (l *tesseractLoader) WithTesseractPath(tesseractPath string) *tesseractLoader {
	l.tesseractPath = tesseractPath
	return l
}

func (l *tesseractLoader) WithTextSplitter(textSplitter TextSplitter) *tesseractLoader {
	l.loader.textSplitter = textSplitter
	return l
}

func (l *tesseractLoader) WithArgs(tesseractArgs []string) *tesseractLoader {
	l.tesseractArgs = tesseractArgs
	return l
}

func (l *tesseractLoader) Load(ctx context.Context) ([]document.Document, error) {

	err := isFile(l.tesseractPath)
	if err != nil {
		return nil, ErrTesseractNotFound
	}

	err = isFile(l.filename)
	if err != nil {
		return nil, err
	}

	documents, err := l.loadFile(ctx)
	if err != nil {
		return nil, err
	}

	if l.loader.textSplitter != nil {
		documents = l.loader.textSplitter.SplitDocuments(documents)
	}

	return documents, nil
}

func (l *tesseractLoader) loadFile(ctx context.Context) ([]document.Document, error) {

	tesseractArgs := []string{l.filename, "stdout"}
	tesseractArgs = append(tesseractArgs, l.tesseractArgs...)

	out, err := exec.CommandContext(ctx, l.tesseractPath, tesseractArgs...).Output()
	if err != nil {
		return nil, err
	}

	metadata := make(types.Meta)
	metadata[SourceMetadataKey] = l.filename

	return []document.Document{
		{
			Content:  string(out),
			Metadata: metadata,
		},
	}, nil
}
