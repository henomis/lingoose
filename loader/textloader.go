package loader

import (
	"fmt"
	"os"

	"github.com/henomis/lingoose/document"
)

const (
	SourceMetadataKey = "source"
)

type TextLoader struct {
	filename string
	metadata map[string]interface{}
}

func NewTextLoader(filename string, metadata map[string]interface{}) (*TextLoader, error) {

	if metadata == nil {
		metadata = make(map[string]interface{})
	} else {
		_, ok := metadata[SourceMetadataKey]
		if ok {
			return nil, fmt.Errorf("metadata key %s is reserved", SourceMetadataKey)
		}
	}

	metadata[SourceMetadataKey] = filename

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
