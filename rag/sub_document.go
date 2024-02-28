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
	llm LLM
}

var SubDocumentRAGSummarizePrompt = "Write a concise summary of the following text:\n\n{{.text}}"

func NewSubDocumentRAG(index *index.Index, llm LLM) *SubDocumentRAG {
	return &SubDocumentRAG{
		RAG: *New(index).
			WithChunkSize(defaultSubDocumentRAGChunkSize).
			WithChunkOverlap(defaultSubDocumentRAGChunkOverlap),
		llm: llm,
	}
}

func (r *SubDocumentRAG) WithChunkSize(chunkSize uint) *SubDocumentRAG {
	r.chunkSize = chunkSize
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
	for _, source := range sources {
		documents, err := r.addSource(ctx, source)
		if err != nil {
			return err
		}

		subDocuments, err := r.generateSubDocuments(ctx, documents)
		if err != nil {
			return err
		}

		err = r.index.LoadFromDocuments(ctx, subDocuments)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *SubDocumentRAG) generateSubDocuments(
	ctx context.Context,
	documents []document.Document,
) ([]document.Document, error) {
	var subDocuments []document.Document

	for _, chunk := range documents {
		t := thread.New().AddMessages(
			thread.NewUserMessage().AddContent(
				thread.NewTextContent(SubDocumentRAGSummarizePrompt).Format(
					types.M{
						"text": chunk.Content,
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
			defaultSubDocumentRAGChildChunkSize,
			0,
		).SplitDocuments(documents)

		for i := range subChunks {
			subChunks[i].Content = summary + "\n" + subChunks[i].Content
		}

		subDocuments = append(subDocuments, subChunks...)
	}

	return subDocuments, nil
}
