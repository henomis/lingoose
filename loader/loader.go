package loader

import (
	"fmt"

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

type loader struct {
	textSplitter TextSplitter
}
