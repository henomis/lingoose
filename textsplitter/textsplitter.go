package textsplitter

import (
	"log"
	"strings"
)

type LenFunction func(string) int

type TextSplitter struct {
	chunkSize      int
	chunkOverlap   int
	lengthFunction LenFunction
}

func (t *TextSplitter) mergeSplits(splits []string, separator string) []string {
	docs := make([]string, 0)
	currentDoc := make([]string, 0)
	total := 0
	for _, d := range splits {
		splitLen := t.lengthFunction(d)

		if total+splitLen+getSLen(currentDoc, separator, 0) > t.chunkSize {
			if total > t.chunkSize {
				log.Printf("Created a chunk of size %d, which is longer than the specified %d", total, t.chunkSize)
			}
			if len(currentDoc) > 0 {
				doc := t.joinDocs(currentDoc, separator)
				if doc != "" {
					docs = append(docs, doc)
				}
				for (total > t.chunkOverlap) || (getSLen(currentDoc, separator, 0) > t.chunkSize) && total > 0 {
					total -= t.lengthFunction(currentDoc[0]) + getSLen(currentDoc, separator, 1)
					currentDoc = currentDoc[1:]
				}
			}
		}
		currentDoc = append(currentDoc, d)
		total += getSLen(currentDoc, separator, 1)
		total += splitLen
	}
	doc := t.joinDocs(currentDoc, separator)
	if doc != "" {
		docs = append(docs, doc)
	}
	return docs
}

func (t *TextSplitter) joinDocs(docs []string, separator string) string {
	text := strings.Join(docs, separator)
	return strings.TrimSpace(text)
}

func getSLen(currentDoc []string, separator string, compareLen int) int {
	if len(currentDoc) > compareLen {
		return len(separator)
	}

	return 0
}
