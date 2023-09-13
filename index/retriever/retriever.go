package retriever

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/index"
	indexoption "github.com/henomis/lingoose/index/option"
)

type Index interface {
	LoadFromDocuments(context.Context, []document.Document) error
	Query(context.Context, string, ...indexoption.Option) (index.SearchResults, error)
}

const (
	defautTopK = 3
)

type Retriever struct {
	index              Index
	topK               int
	documents          *[]document.Document
	areDocumentsLoaded bool
}

func New(index Index, documents *[]document.Document) *Retriever {
	return &Retriever{
		index:     index,
		topK:      defautTopK,
		documents: documents,
	}
}

func (r *Retriever) WithTopK(topK int) *Retriever {
	r.topK = topK
	return r
}

func (r *Retriever) Query(ctx context.Context, query string) ([]document.Document, error) {

	if r.documents == nil {
		return nil, fmt.Errorf("documents are not defined")
	}

	if !r.areDocumentsLoaded {
		err := r.index.LoadFromDocuments(context.Background(), *r.documents)
		if err != nil {
			return nil, err
		}
		r.areDocumentsLoaded = true
	}

	results, err := r.index.Query(context.Background(), query, indexoption.WithTopK(r.topK))
	if err != nil {
		return nil, err
	}

	return results.ToDocuments(), nil
}
