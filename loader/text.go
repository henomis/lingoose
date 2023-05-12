package loader

import (
	"context"
	"fmt"
	"os"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
)

type textLoader struct {
	loader loader

	filename string
	metadata types.Meta
}

func NewTextLoader(filename string, metadata types.Meta) *textLoader {
	return &textLoader{
		filename: filename,
		metadata: metadata,
	}
}

func (t *textLoader) WithTextSplitter(textSplitter TextSplitter) *textLoader {
	t.loader.textSplitter = textSplitter
	return t
}

func (t *textLoader) Load(ctx context.Context) ([]document.Document, error) {

	err := t.validate()
	if err != nil {
		return nil, err
	}

	text, err := os.ReadFile(t.filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	documents := []document.Document{
		{
			Content:  string(text),
			Metadata: t.metadata,
		},
	}

	if t.loader.textSplitter != nil {
		documents = t.loader.textSplitter.SplitDocuments(documents)
	}

	return documents, nil
}

func (t *textLoader) validate() error {
	if t.metadata == nil {
		t.metadata = make(types.Meta)
	} else {
		_, ok := t.metadata[SourceMetadataKey]
		if ok {
			return fmt.Errorf("%s: metadata key %s is reserved", ErrorInternal, SourceMetadataKey)
		}
	}

	t.metadata[SourceMetadataKey] = t.filename

	fileStat, err := os.Stat(t.filename)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	if fileStat.IsDir() {
		return fmt.Errorf("%s: %w", ErrorInternal, os.ErrNotExist)
	}

	return nil
}
