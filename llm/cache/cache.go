package cache

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/index"
	indexoption "github.com/henomis/lingoose/index/option"
	"github.com/henomis/lingoose/types"
)

var ErrCacheMiss = fmt.Errorf("cache miss")

const (
	defaultTopK            = 1
	defaultScoreThreshold  = 0.9
	cacheAnswerMetadataKey = "cache-answer"
)

type Cache struct {
	embedder       index.Embedder
	index          *index.Index
	topK           int
	scoreThreshold float64
}

type Result struct {
	Answer    []string
	Embedding []float64
}

func New(index *index.Index) *Cache {
	return &Cache{
		embedder:       index.Embedder(),
		index:          index,
		topK:           defaultTopK,
		scoreThreshold: defaultScoreThreshold,
	}
}

func (c *Cache) WithTopK(topK int) *Cache {
	c.topK = topK
	return c
}

func (c *Cache) WithScoreThreshold(scoreThreshold float64) *Cache {
	c.scoreThreshold = scoreThreshold
	return c
}

func (c *Cache) Get(ctx context.Context, query string) (*Result, error) {
	embedding, err := c.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}

	results, err := c.index.Search(ctx, embedding[0], indexoption.WithTopK(c.topK))
	if err != nil {
		return nil, err
	}

	answers, cacheHit := c.extractResults(results)
	if cacheHit {
		return &Result{
			Answer:    answers,
			Embedding: embedding[0],
		}, nil
	}

	return &Result{Embedding: embedding[0]}, ErrCacheMiss
}

func (c *Cache) Set(ctx context.Context, embedding []float64, answer string) error {
	return c.index.Add(ctx, &index.Data{
		Values: embedding,
		Metadata: types.Meta{
			cacheAnswerMetadataKey: answer,
		},
	})
}

func (c *Cache) Clear(ctx context.Context) error {
	return c.index.Drop(ctx)
}

func (c *Cache) extractResults(results index.SearchResults) ([]string, bool) {
	var output []string

	for _, result := range results {
		if result.Score > c.scoreThreshold {
			answer, ok := result.Metadata[cacheAnswerMetadataKey]
			if !ok {
				continue
			}

			output = append(output, answer.(string))
		}
	}

	return output, len(output) > 0
}
