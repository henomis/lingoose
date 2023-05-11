package loader

import (
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
)

type pdfLoader struct {
	loader loader

	pdfToTextPath string
	path          string
}

func NewPDFToTextLoader(pdfToTextPath, path string) *pdfLoader {
	return &pdfLoader{
		pdfToTextPath: pdfToTextPath,
		path:          path,
	}
}

func (p *pdfLoader) WithTextSplitter(textSplitter TextSplitter) *pdfLoader {
	p.loader.textSplitter = textSplitter
	return p
}

func (p *pdfLoader) Load() ([]document.Document, error) {

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
		documents, err = p.loadDir()
	} else {
		documents, err = p.loadFile()
	}
	if err != nil {
		return nil, err
	}

	if p.loader.textSplitter != nil {
		documents = p.loader.textSplitter.SplitDocuments(documents)
	}

	return documents, nil
}

func (p *pdfLoader) loadFile() ([]document.Document, error) {
	out, err := exec.Command(p.pdfToTextPath, p.path, "-").Output()
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

func (p *pdfLoader) loadDir() ([]document.Document, error) {
	docs := []document.Document{}

	err := filepath.Walk(p.path, func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(info.Name(), ".pdf") {

			d, err := NewPDFToTextLoader(p.pdfToTextPath, path).loadFile()
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
