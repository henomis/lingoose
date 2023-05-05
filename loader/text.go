package loader

import (
	"fmt"
	"os"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
)

var (
	ErrorInternal = fmt.Errorf("internal error")
)

const (
	SourceMetadataKey = "source"
)

type textLoader struct {
	filename string
	metadata types.Meta
}

func NewTextLoader(filename string, metadata types.Meta) (*textLoader, error) {

	if metadata == nil {
		metadata = make(types.Meta)
	} else {
		_, ok := metadata[SourceMetadataKey]
		if ok {
			return nil, fmt.Errorf("%s: metadata key %s is reserved", ErrorInternal, SourceMetadataKey)
		}
	}

	metadata[SourceMetadataKey] = filename

	fileStat, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	if fileStat.IsDir() {
		return nil, fmt.Errorf("%s: %w", ErrorInternal, os.ErrNotExist)
	}

	return &textLoader{
		filename: filename,
		metadata: metadata,
	}, nil
}

func (t *textLoader) Load() ([]document.Document, error) {
	text, err := os.ReadFile(t.filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	return []document.Document{
		{
			Content:  string(text),
			Metadata: t.metadata,
		},
	}, nil
}
