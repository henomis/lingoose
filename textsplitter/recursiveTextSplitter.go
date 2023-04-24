package textsplitter

import (
	"strings"

	"github.com/henomis/lingoose/document"
)

type RecursiveCharacterTextSplitter struct {
	textSplitter
	separators []string
}

func NewRecursiveCharacterTextSplitter(separators []string, chunkSize int, chunkOverlap int, lengthFunction LenFunction) *RecursiveCharacterTextSplitter {

	if lengthFunction == nil {
		lengthFunction = func(s string) int {
			return len(s)
		}
	}

	if len(separators) == 0 {
		separators = []string{"\n\n", "\n", " ", ""}
	}

	return &RecursiveCharacterTextSplitter{
		textSplitter: textSplitter{
			chunkSize:      chunkSize,
			chunkOverlap:   chunkOverlap,
			lengthFunction: lengthFunction,
		},
		separators: separators,
	}

}

func (t *RecursiveCharacterTextSplitter) SplitDocuments(documents []document.Document) []document.Document {

	docs := make([]document.Document, 0)

	for i, doc := range documents {
		for _, chunk := range t.SplitText(doc.Content) {
			docs = append(docs,
				document.Document{
					Content:  chunk,
					Metadata: documents[i].Metadata,
				},
			)
		}
	}

	return docs

}

func (t *RecursiveCharacterTextSplitter) SplitText(text string) []string {
	// Split incoming text and return chunks.
	finalChunks := []string{}
	// Get appropriate separator to use
	separator := t.separators[len(t.separators)-1]
	for _, s := range t.separators {
		if s == "" {
			separator = s
			break
		}
		if strings.Contains(text, s) {
			separator = s
			break
		}
	}
	// Now that we have the separator, split the text
	var splits []string
	if separator != "" {
		splits = strings.Split(text, separator)
	} else {
		splits = strings.Split(text, "")
	}
	// Now go merging things, recursively splitting longer texts.
	goodSplits := []string{}
	for _, s := range splits {
		if t.lengthFunction(s) < t.chunkSize {
			goodSplits = append(goodSplits, s)
		} else {
			if len(goodSplits) > 0 {
				mergedText := t.mergeSplits(goodSplits, separator)
				finalChunks = append(finalChunks, mergedText...)
				goodSplits = []string{}
			}
			otherInfo := t.SplitText(s)
			finalChunks = append(finalChunks, otherInfo...)
		}
	}
	if len(goodSplits) > 0 {
		mergedText := t.mergeSplits(goodSplits, separator)
		finalChunks = append(finalChunks, mergedText...)
	}
	return finalChunks
}
