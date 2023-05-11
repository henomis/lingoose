package loader

import "github.com/henomis/lingoose/document"

type TextSplitter interface {
	SplitDocuments(documents []document.Document) []document.Document
}

type loader struct {
	textSplitter TextSplitter
}
