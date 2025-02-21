package loader

import (
	"fmt"
	"os"

	"github.com/rsest/lingoose/document"
)

var (
	ErrInternal = fmt.Errorf("internal error")
)

const (
	SourceMetadataKey = "source"
)

type TextSplitter interface {
	SplitDocuments(documents []document.Document) []document.Document
}

type Loader struct {
	textSplitter TextSplitter
}

func isFile(filename string) error {
	fileStat, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInternal, err)
	}

	if fileStat.IsDir() {
		return fmt.Errorf("%w: %w", ErrInternal, os.ErrNotExist)
	}

	return nil
}
