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

type PDFLoader struct {
	loader Loader

	pdfToTextPath string
	path          string
}

func NewPDFToTextLoader(path string) *PDFLoader {
	return &PDFLoader{
		pdfToTextPath: defaultPdfToTextPath,
		path:          path,
	}
}

func (p *PDFLoader) WithPDFToTextPath(pdfToTextPath string) *PDFLoader {
	p.pdfToTextPath = pdfToTextPath
	return p
}

func (p *PDFLoader) WithTextSplitter(textSplitter TextSplitter) *PDFLoader {
	p.loader.textSplitter = textSplitter
	return p
}

func (p *PDFLoader) Load(ctx context.Context) ([]document.Document, error) {
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

func (p *PDFLoader) loadFile(ctx context.Context) ([]document.Document, error) {
	//nolint:gosec
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

func (p *PDFLoader) loadDir(ctx context.Context) ([]document.Document, error) {
	docs := []document.Document{}

	err := filepath.Walk(p.path, func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(info.Name(), ".pdf") {
			d, errLoad := NewPDFToTextLoader(path).WithPDFToTextPath(p.pdfToTextPath).loadFile(ctx)
			if errLoad != nil {
				return errLoad
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
