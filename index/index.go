package index

import "github.com/henomis/lingoose/document"

type SearchResponse struct {
	Document document.Document
	Score    float32
	Index    int
}
