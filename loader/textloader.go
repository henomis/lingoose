package loader

import (
	"os"

	"github.com/henomis/lingoose/document"
)

type TextLoader struct {
	filename string
	metadata map[string]interface{}
}

func NewTextLoader(filename string, metadata map[string]interface{}) (*TextLoader, error) {

	fileStat, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	if fileStat.IsDir() {
		return nil, os.ErrNotExist
	}

	return &TextLoader{
		filename: filename,
		metadata: metadata,
	}, nil
}

func (t *TextLoader) Load() ([]document.Document, error) {
	text, err := os.ReadFile(t.filename)
	if err != nil {
		return nil, err
	}

	return []document.Document{
		{
			Content:  string(text),
			Metadata: t.metadata,
		},
	}, nil
}
