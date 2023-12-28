package rag

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/option"
	"github.com/henomis/lingoose/loader"
	"github.com/henomis/lingoose/textsplitter"
	"github.com/henomis/lingoose/thread"
)

const (
	defaultChunkSize    = 1000
	defaultChunkOverlap = 0
	defaultTopK         = 1
)

type LLM interface {
	Generate(context.Context, *thread.Thread) error
}

type RAG struct {
	index        *index.Index
	chunkSize    uint
	chunkOverlap uint
	topK         uint
}

type RAGFusion struct {
	RAG
	llm LLM
}

func New(index *index.Index) *RAG {
	return &RAG{
		index:        index,
		chunkSize:    defaultChunkSize,
		chunkOverlap: defaultChunkOverlap,
		topK:         defaultTopK,
	}
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

func (r *RAG) AddFiles(ctx context.Context, filePath ...string) error {
	for _, f := range filePath {
		documents, err := r.addFile(ctx, f)
		if err != nil {
			return err
		}

		err = r.index.LoadFromDocuments(ctx, documents)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RAG) Retrieve(ctx context.Context, query string) ([]index.SearchResult, error) {
	results, err := r.index.Query(ctx, query, option.WithTopK(int(r.topK)))
	return results, err
}

func (r *RAG) addFile(ctx context.Context, filePath string) ([]document.Document, error) {
	var documents []document.Document
	var err error
	switch filepath.Ext(filePath) {
	case ".pdf":
		documents, err = loader.NewPDFToTextLoader(filePath).Load(ctx)
	case ".docx":
		documents, err = loader.NewLibreOfficeLoader(filePath).Load(ctx)
	case ".txt":
		documents, err = loader.NewTextLoader(filePath, nil).Load(ctx)
	default:
		return nil, fmt.Errorf("unsupported file type")
	}

	if err != nil {
		return nil, err
	}

	return textsplitter.NewRecursiveCharacterTextSplitter(
		int(r.chunkSize),
		int(r.chunkOverlap),
	).SplitDocuments(documents), nil
}
