package rag

import (
	"context"
	"regexp"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/textsplitter"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

const (
	defaultSubDocumentRAGChunkSize      = 8192
	defaultSubDocumentRAGChunkOverlap   = 0
	defaultSubDocumentRAGChildChunkSize = 512
)

type SubDocumentRAG struct {
	RAG
	childChunkSize uint
	llm            LLM
}

//nolint:lll
var SubDocumentRAGSummarizePrompt = "Please give a concise summary of the context in 1-2 sentences.\n\nContext: {{.context}}"

func NewSubDocument(index *index.Index, llm LLM) *SubDocumentRAG {
	return &SubDocumentRAG{
		RAG: *New(index).
			WithChunkSize(defaultSubDocumentRAGChunkSize).
			WithChunkOverlap(defaultSubDocumentRAGChunkOverlap),
		childChunkSize: defaultSubDocumentRAGChildChunkSize,
		llm:            llm,
	}
}

func (r *SubDocumentRAG) WithChunkSize(chunkSize uint) *SubDocumentRAG {
	r.chunkSize = chunkSize
	return r
}

func (r *SubDocumentRAG) WithChildChunkSize(childChunkSize uint) *SubDocumentRAG {
	r.childChunkSize = childChunkSize
	return r
}

func (r *SubDocumentRAG) WithChunkOverlap(chunkOverlap uint) *SubDocumentRAG {
	r.chunkOverlap = chunkOverlap
	return r
}

func (r *SubDocumentRAG) WithTopK(topK uint) *SubDocumentRAG {
	r.topK = topK
	return r
}

func (r *SubDocumentRAG) WithLoader(sourceRegexp *regexp.Regexp, loader Loader) *SubDocumentRAG {
	r.loaders[sourceRegexp] = loader
	return r
}

func (r *SubDocumentRAG) AddSources(ctx context.Context, sources ...string) error {
	ctx, span, err := r.startObserveSpan(
		ctx,
		"rag-subdocument-add-sources",
		types.M{
			"chunkSize":      r.chunkSize,
			"childChunkSize": r.childChunkSize,
			"chunkOverlap":   r.chunkOverlap,
		},
	)
	if err != nil {
		return err
	}

	for _, source := range sources {
		documents, errAddSource := r.addSource(ctx, source)
		if errAddSource != nil {
			return errAddSource
		}

		subDocuments, errAddSource := r.generateSubDocuments(ctx, documents)
		if errAddSource != nil {
			return errAddSource
		}

		errAddSource = r.index.LoadFromDocuments(ctx, subDocuments)
		if errAddSource != nil {
			return errAddSource
		}
	}

	err = r.stopObserveSpan(ctx, span)
	if err != nil {
		return err
	}

	return nil
}

func (r *SubDocumentRAG) generateSubDocuments(
	ctx context.Context,
	documents []document.Document,
) ([]document.Document, error) {
	var subDocuments []document.Document

	for _, doc := range documents {
		t := thread.New().AddMessages(
			thread.NewUserMessage().AddContent(
				thread.NewTextContent(SubDocumentRAGSummarizePrompt).Format(
					types.M{
						"context": doc.Content,
					},
				),
			),
		)

		err := r.llm.Generate(ctx, t)
		if err != nil {
			return nil, err
		}
		summary := t.LastMessage().Contents[0].AsString()

		subChunks := textsplitter.NewRecursiveCharacterTextSplitter(
			int(r.childChunkSize),
			0,
		).SplitDocuments([]document.Document{doc})

		for i := range subChunks {
			subChunks[i].Content = summary + "\n" + subChunks[i].Content
		}

		subDocuments = append(subDocuments, subChunks...)
	}

	return subDocuments, nil
}
