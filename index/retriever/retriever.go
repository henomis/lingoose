package retriever

import (
	"context"

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
	index Index
	topK  int
}

func New(index Index) *Retriever {
	return &Retriever{
		index: index,
		topK:  defautTopK,
	}
}

func (r *Retriever) WithTopK(topK int) *Retriever {
	r.topK = topK
	return r
}

func (r *Retriever) Query(ctx context.Context, query string, documents []document.Document) ([]document.Document, error) {
	err := r.index.LoadFromDocuments(context.Background(), documents)
	if err != nil {
		return nil, err
	}

	results, err := r.index.Query(context.Background(), query, indexoption.WithTopK(r.topK))
	if err != nil {
		return nil, err
	}

	return results.ToDocuments(), nil
}
