package loader

import (
	"fmt"
	"os"

	"github.com/henomis/lingoose/document"
)

var (
	ErrorInternal = fmt.Errorf("internal error")
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
		return fmt.Errorf("%s: %w", ErrorInternal, err)
	}

	if fileStat.IsDir() {
		return fmt.Errorf("%s: %w", ErrorInternal, os.ErrNotExist)
	}

	return nil
}
