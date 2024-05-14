package rag

import (
	"context"
	"fmt"
	"regexp"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/option"
	"github.com/henomis/lingoose/loader"
	obs "github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/textsplitter"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

const (
	defaultChunkSize    = 1000
	defaultChunkOverlap = 0
	defaultTopK         = 1
)

type LLM interface {
	Generate(context.Context, *thread.Thread) error
}

type Loader interface {
	LoadFromSource(context.Context, string) ([]document.Document, error)
}

type observer interface {
	Span(s *obs.Span) (*obs.Span, error)
	SpanEnd(s *obs.Span) (*obs.Span, error)
}

type RAG struct {
	index           *index.Index
	chunkSize       uint
	chunkOverlap    uint
	topK            uint
	loaders         map[*regexp.Regexp]Loader // this map a regexp as string to a loader
	observer        observer
	observerTraceID string
}

func New(index *index.Index) *RAG {
	rag := &RAG{
		index:        index,
		chunkSize:    defaultChunkSize,
		chunkOverlap: defaultChunkOverlap,
		topK:         defaultTopK,
		loaders:      make(map[*regexp.Regexp]Loader),
	}

	return rag.withDefaultLoaders()
}

func (r *RAG) WithChunkSize(chunkSize uint) *RAG {
	r.chunkSize = chunkSize
	return r
}

func (r *RAG) WithChunkOverlap(chunkOverlap uint) *RAG {
	r.chunkOverlap = chunkOverlap
	return r
}

func (r *RAG) WithTopK(topK uint) *RAG {
	r.topK = topK
	return r
}

func (r *RAG) withDefaultLoaders() *RAG {
	r.loaders[regexp.MustCompile(`.*\.pdf`)] = loader.NewPDFToText()
	r.loaders[regexp.MustCompile(`.*\.docx`)] = loader.NewLibreOffice()
	r.loaders[regexp.MustCompile(`.*\.txt`)] = loader.NewText()

	return r
}

func (r *RAG) WithLoader(sourceRegexp *regexp.Regexp, loader Loader) *RAG {
	r.loaders[sourceRegexp] = loader
	return r
}

func (r *RAG) WithObserver(observer observer, traceID string) *RAG {
	r.observer = observer
	r.observerTraceID = traceID
	return r
}

func (r *RAG) AddSources(ctx context.Context, sources ...string) error {
	var err error
	var span *obs.Span
	if r.observer != nil {
		span, err = r.startObserveSpan(
			ctx,
			"RAG AddSources",
			types.M{
				"chunkSize":    r.chunkSize,
				"chunkOverlap": r.chunkOverlap,
			},
		)
		if err != nil {
			return err
		}
		ctx = obs.ContextWithParentID(ctx, span.ID)
	}

	for _, source := range sources {
		documents, errAddSource := r.addSource(ctx, source)
		if errAddSource != nil {
			return errAddSource
		}

		errAddSource = r.index.LoadFromDocuments(ctx, documents)
		if errAddSource != nil {
			return errAddSource
		}
	}

	if r.observer != nil {
		err = r.stopObserveSpan(span)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RAG) AddDocuments(ctx context.Context, documents ...document.Document) error {
	var err error
	var span *obs.Span
	if r.observer != nil {
		span, err = r.startObserveSpan(
			ctx,
			"RAG AddDocument",
			types.M{
				"chunkSize":    r.chunkSize,
				"chunkOverlap": r.chunkOverlap,
			},
		)
		if err != nil {
			return err
		}
		ctx = obs.ContextWithParentID(ctx, span.ID)
	}

	err = r.index.LoadFromDocuments(ctx, documents)
	if err != nil {
		return err
	}

	if r.observer != nil {
		err = r.stopObserveSpan(span)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RAG) Retrieve(ctx context.Context, query string) ([]string, error) {
	var err error
	var span *obs.Span
	if r.observer != nil {
		span, err = r.startObserveSpan(
			ctx,
			"RAG Retrieve",
			types.M{
				"query": query,
				"topK":  r.topK,
			},
		)
		if err != nil {
			return nil, err
		}
		ctx = obs.ContextWithParentID(ctx, span.ID)
	}

	texts, err := r.retrieve(ctx, query)
	if err != nil {
		return nil, err
	}

	if r.observer != nil {
		err = r.stopObserveSpan(span)
		if err != nil {
			return nil, err
		}
	}

	return texts, nil
}

func (r *RAG) retrieve(ctx context.Context, query string) ([]string, error) {
	results, err := r.index.Query(ctx, query, option.WithTopK(int(r.topK)))
	var resultsAsString []string
	for _, result := range results {
		resultsAsString = append(resultsAsString, result.Content())
	}

	return resultsAsString, err
}

func (r *RAG) addSource(ctx context.Context, source string) ([]document.Document, error) {
	var sourceLoader Loader
	for regexpStr, loader := range r.loaders {
		if regexpStr.MatchString(source) {
			sourceLoader = loader
		}
	}

	if sourceLoader == nil {
		return nil, fmt.Errorf("unsupported source type")
	}

	documents, err := sourceLoader.LoadFromSource(ctx, source)
	if err != nil {
		return nil, err
	}

	return textsplitter.NewRecursiveCharacterTextSplitter(
		int(r.chunkSize),
		int(r.chunkOverlap),
	).SplitDocuments(documents), nil
}

func (r *RAG) startObserveSpan(ctx context.Context, name string, input any) (*obs.Span, error) {
	return r.observer.Span(
		&obs.Span{
			TraceID:  r.observerTraceID,
			ParentID: obs.ContextValueParentID(ctx),
			Name:     name,
			Input:    input,
		},
	)
}

func (r *RAG) stopObserveSpan(span *obs.Span) error {
	_, err := r.observer.SpanEnd(span)
	return err
}
