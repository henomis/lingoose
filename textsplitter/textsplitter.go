package textsplitter

import (
	"log"
	"strings"
)

type LenFunction func(string) int

type textSplitter struct {
	chunkSize      int
	chunkOverlap   int
	lengthFunction LenFunction
}

func (p *textSplitter) mergeSplits(splits []string, separator string) []string {

	docs := make([]string, 0)
	current_doc := make([]string, 0)
	total := 0
	for _, d := range splits {
		splitLen := p.lengthFunction(d)

		if total+splitLen+getSLen(current_doc, separator, 0) > p.chunkSize {
			if total > p.chunkSize {
				log.Printf("Created a chunk of size %d, which is longer than the specified %d", total, p.chunkSize)
			}
			if len(current_doc) > 0 {
				doc := p.joinDocs(current_doc, separator)
				if doc != "" {
					docs = append(docs, doc)
				}
				for (total > p.chunkOverlap) || (getSLen(current_doc, separator, 0) > p.chunkSize) && total > 0 {
					total -= p.lengthFunction(current_doc[0]) + getSLen(current_doc, separator, 1)
					current_doc = current_doc[1:]
				}
			}
		}
		current_doc = append(current_doc, d)
		total += getSLen(current_doc, separator, 1)
		total += splitLen
	}
	doc := p.joinDocs(current_doc, separator)
	if doc != "" {
		docs = append(docs, doc)
	}
	return docs
}

func (t *textSplitter) joinDocs(docs []string, separator string) string {
	text := strings.Join(docs, separator)
	return strings.TrimSpace(text)
}

func getSLen(current_doc []string, separator string, compareLen int) int {
	if len(current_doc) > compareLen {
		return len(separator)
	}

	return 0
}
