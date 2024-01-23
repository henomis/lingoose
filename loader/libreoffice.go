package loader

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
)

var (
	ErrLibreOfficeNotFound = fmt.Errorf("pdftotext not found")
	defaultLibreOfficePath = "/usr/bin/soffice"
)

type LibreOfficeLoader struct {
	loader Loader

	libreOfficePath string
	libreOfficeArgs []string
	filename        string
}

func NewLibreOfficeLoader(filename string) *LibreOfficeLoader {
	return &LibreOfficeLoader{
		libreOfficePath: defaultLibreOfficePath,
		libreOfficeArgs: []string{"--headless", "--convert-to", "txt:Text", "--cat"},
		filename:        filename,
	}
}

func (l *LibreOfficeLoader) WithLibreOfficePath(libreOfficePath string) *LibreOfficeLoader {
	l.libreOfficePath = libreOfficePath
	return l
}

func (l *LibreOfficeLoader) WithTextSplitter(textSplitter TextSplitter) *LibreOfficeLoader {
	l.loader.textSplitter = textSplitter
	return l
}

func (l *LibreOfficeLoader) WithArgs(libreOfficeArgs []string) *LibreOfficeLoader {
	l.libreOfficeArgs = libreOfficeArgs
	return l
}

func (l *LibreOfficeLoader) Load(ctx context.Context) ([]document.Document, error) {
	err := isFile(l.libreOfficePath)
	if err != nil {
		return nil, ErrLibreOfficeNotFound
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

func (l *LibreOfficeLoader) LoadFromSource(ctx context.Context, source string) ([]document.Document, error) {
	l.filename = source
	return l.Load(ctx)
}

func (l *LibreOfficeLoader) loadFile(ctx context.Context) ([]document.Document, error) {
	libreOfficeArgs := append(l.libreOfficeArgs, l.filename)

	//nolint:gosec
	out, err := exec.CommandContext(ctx, l.libreOfficePath, libreOfficeArgs...).Output()
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
