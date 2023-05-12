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

type libreOfficeLoader struct {
	loader loader

	libreOfficePath string
	libreOfficeArgs []string
	filename        string
}

func NewLibreOfficeLoader(filename string) *libreOfficeLoader {
	return &libreOfficeLoader{
		libreOfficePath: defaultLibreOfficePath,
		libreOfficeArgs: []string{"--headless", "--convert-to", "txt:Text", "--cat"},
		filename:        filename,
	}
}

func (l *libreOfficeLoader) WithLibreOfficePath(libreOfficePath string) *libreOfficeLoader {
	l.libreOfficePath = libreOfficePath
	return l
}

func (l *libreOfficeLoader) WithTextSplitter(textSplitter TextSplitter) *libreOfficeLoader {
	l.loader.textSplitter = textSplitter
	return l
}

func (l *libreOfficeLoader) WithArgs(libreOfficeArgs []string) *libreOfficeLoader {
	l.libreOfficeArgs = libreOfficeArgs
	return l
}

func (l *libreOfficeLoader) Load(ctx context.Context) ([]document.Document, error) {

	err := isFile(l.filename)
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

func (l *libreOfficeLoader) loadFile(ctx context.Context) ([]document.Document, error) {

	libreOfficeArgs := append(l.libreOfficeArgs, l.filename)

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
