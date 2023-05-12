package loader

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
)

var (
	ErrPdfToTextNotFound = fmt.Errorf("pdftotext not found")
	defaultPdfToTextPath = "/usr/bin/pdftotext"
)

type pdfLoader struct {
	loader loader

	pdfToTextPath string
	path          string
}

func NewPDFToTextLoader(path string) *pdfLoader {
	return &pdfLoader{
		pdfToTextPath: defaultPdfToTextPath,
		path:          path,
	}
}

func (p *pdfLoader) WithPDFToTextPath(pdfToTextPath string) *pdfLoader {
	p.pdfToTextPath = pdfToTextPath
	return p
}

func (p *pdfLoader) WithTextSplitter(textSplitter TextSplitter) *pdfLoader {
	p.loader.textSplitter = textSplitter
	return p
}

func (p *pdfLoader) Load(ctx context.Context) ([]document.Document, error) {

	_, err := os.Stat(p.pdfToTextPath)
	if err != nil {
		return nil, ErrPdfToTextNotFound
	}

	fileInfo, err := os.Stat(p.path)
	if err != nil {
		return nil, err
	}

	var documents []document.Document
	if fileInfo.IsDir() {
		documents, err = p.loadDir(ctx)
	} else {
		documents, err = p.loadFile(ctx)
	}
	if err != nil {
		return nil, err
	}

	if p.loader.textSplitter != nil {
		documents = p.loader.textSplitter.SplitDocuments(documents)
	}

	return documents, nil
}

func (p *pdfLoader) loadFile(ctx context.Context) ([]document.Document, error) {
	out, err := exec.CommandContext(ctx, p.pdfToTextPath, p.path, "-").Output()
	if err != nil {
		return nil, err
	}

	metadata := make(types.Meta)
	metadata[SourceMetadataKey] = p.path

	return []document.Document{
		{
			Content:  string(out),
			Metadata: metadata,
		},
	}, nil
}

func (p *pdfLoader) loadDir(ctx context.Context) ([]document.Document, error) {
	docs := []document.Document{}

	err := filepath.Walk(p.path, func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(info.Name(), ".pdf") {

			d, err := NewPDFToTextLoader(path).WithPDFToTextPath(p.pdfToTextPath).loadFile(ctx)
			if err != nil {
				return err
			}

			docs = append(docs, d...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return docs, nil
}
