package rag

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/rsest/lingoose/index"
	"github.com/rsest/lingoose/index/option"
	obs "github.com/rsest/lingoose/observer"
	"github.com/rsest/lingoose/thread"
	"github.com/rsest/lingoose/types"
)

var ragFusionPrompts = []string{
	"You are a helpful assistant that generates multiple search queries based on a single input query.",
	"Generate multiple search queries related to: %s",
	"OUTPUT (4 queries):",
}

type Fusion struct {
	RAG
	llm LLM
}

func NewFusion(index *index.Index, llm LLM) *Fusion {
	return &Fusion{
		RAG: *New(index),
		llm: llm,
	}
}

func (r *Fusion) Retrieve(ctx context.Context, query string) ([]string, error) {
	ctx, span, err := r.startObserveSpan(
		ctx,
		"rag-fusion-retrieve",
		types.M{
			"query": query,
			"topK":  r.topK,
		},
	)
	if err != nil {
		return nil, err
	}
	ctx = obs.ContextWithParentID(ctx, span.ID)

	texts, err := r.retrieve(ctx, query)
	if err != nil {
		return nil, err
	}

	err = r.stopObserveSpan(ctx, span)
	if err != nil {
		return nil, err
	}

	return texts, nil
}

func (r *Fusion) retrieve(ctx context.Context, query string) ([]string, error) {
	if r.llm == nil {
		return nil, fmt.Errorf("llm is not set")
	}

	t := thread.New().AddMessages(
		thread.NewSystemMessage().AddContent(
			thread.NewTextContent(
				ragFusionPrompts[0],
			),
		),
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(
				fmt.Sprintf(ragFusionPrompts[1], query),
			),
		),
		thread.NewUserMessage().AddContent(
			thread.NewTextContent(
				ragFusionPrompts[2],
			),
		),
	)

	err := r.llm.Generate(ctx, t)
	if err != nil {
		return nil, err
	}

	lastMessage := t.LastMessage()
	content, _ := lastMessage.Contents[0].Data.(string)
	content = strings.TrimSpace(content)
	questions := strings.Split(content, "\n")

	var results index.SearchResults
	for _, question := range questions {
		res, queryErr := r.index.Query(ctx, question, option.WithTopK(int(r.topK)))
		if queryErr != nil {
			return nil, queryErr
		}

		results = append(results, res...)
	}

	return reciprocalRankFusion(results), nil
}

func reciprocalRankFusion(searchResults index.SearchResults) []string {
	const k = 60.0
	searchResultsScoreMap := make(map[string]float64)
	for _, result := range searchResults {
		if _, ok := searchResultsScoreMap[result.ID]; !ok {
			searchResultsScoreMap[result.ID] = 0
		}
		searchResultsScoreMap[result.ID] += 1 / (result.Score + k)
	}

	for i, searchResult := range searchResults {
		searchResults[i].Score = searchResultsScoreMap[searchResult.ID]
	}

	//remove duplicates
	seen := make(map[string]bool)
	var uniqueSearchResults index.SearchResults
	for _, searchResult := range searchResults {
		if _, ok := seen[searchResult.Content()]; !ok {
			uniqueSearchResults = append(uniqueSearchResults, searchResult)
			seen[searchResult.Content()] = true
		}
	}

	//sort by score
	sort.Slice(uniqueSearchResults, func(i, j int) bool {
		return uniqueSearchResults[i].Score > uniqueSearchResults[j].Score
	})

	var results []string
	for _, searchResult := range uniqueSearchResults {
		results = append(results, searchResult.Content())
	}

	return results
}
