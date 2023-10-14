package jsondb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/embedder"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/option"
	"github.com/henomis/lingoose/types"
)

type data struct {
	ID       string     `json:"id"`
	Metadata types.Meta `json:"metadata"`
	Values   []float64  `json:"values"`
}

type DB struct {
	data   []data
	dbPath string
}

type FilterFn func([]index.SearchResult) []index.SearchResult

func New(dbPath string) *DB {
	index := &DB{
		data:   []data{},
		dbPath: dbPath,
	}

	return index
}

func (i DB) save() error {
	jsonContent, err := json.Marshal(i.data)
	if err != nil {
		return err
	}

	return os.WriteFile(i.dbPath, jsonContent, 0600)
}

func (i *DB) load() error {
	if len(i.data) > 0 {
		return nil
	}

	if _, err := os.Stat(i.dbPath); os.IsNotExist(err) {
		return i.save()
	}

	content, err := os.ReadFile(i.dbPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(content, &i.data)
}

func (i *DB) IsEmpty(_ context.Context) (bool, error) {
	err := i.load()
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return len(i.data) == 0, nil
}

func (i *DB) Insert(ctx context.Context, datas []index.Data) error {
	_ = ctx
	err := i.load()
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	var records []data
	for _, item := range datas {
		if item.ID == "" {
			id, errUUID := uuid.NewUUID()
			if errUUID != nil {
				return errUUID
			}
			item.ID = id.String()
		}

		point := data{
			ID:       item.ID,
			Values:   item.Values,
			Metadata: item.Metadata,
		}
		records = append(records, point)
	}

	i.data = append(i.data, records...)

	return i.save()
}

func (i *DB) Search(ctx context.Context, values []float64, options *option.Options) (index.SearchResults, error) {
	err := i.load()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return i.similaritySearch(ctx, values, options)
}

func (i *DB) similaritySearch(
	ctx context.Context,
	embedding embedder.Embedding,
	opts *option.Options,
) (index.SearchResults, error) {
	_ = ctx
	scores, err := i.cosineSimilarityBatch(embedding)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	searchResults := make([]index.SearchResult, len(scores))

	for j, score := range scores {
		searchResults[j] = index.SearchResult{
			Data: index.Data{
				ID:       i.data[j].ID,
				Values:   i.data[j].Values,
				Metadata: i.data[j].Metadata,
			},
			Score: score,
		}
	}

	if opts.Filter != nil {
		searchResults = opts.Filter.(FilterFn)(searchResults)
	}

	return filterSearchResults(searchResults, opts.TopK), nil
}

func (i *DB) cosineSimilarity(a []float64, b []float64) (cosine float64, err error) {
	var count int
	lengthA := len(a)
	lengthB := len(b)
	if lengthA > lengthB {
		count = lengthA
	} else {
		count = lengthB
	}
	sumA := 0.0
	s1 := 0.0
	s2 := 0.0
	for k := 0; k < count; k++ {
		if k >= lengthA {
			s2 += math.Pow(b[k], 2)
			continue
		}
		if k >= lengthB {
			s1 += math.Pow(a[k], 2)
			continue
		}
		sumA += a[k] * b[k]
		s1 += math.Pow(a[k], 2)
		s2 += math.Pow(b[k], 2)
	}
	if s1 == 0 || s2 == 0 {
		return 0.0, errors.New("vectors should not be null (all zeros)")
	}
	return sumA / (math.Sqrt(s1) * math.Sqrt(s2)), nil
}

func (i *DB) cosineSimilarityBatch(a embedder.Embedding) ([]float64, error) {
	var err error
	scores := make([]float64, len(i.data))

	for j := range i.data {
		scores[j], err = i.cosineSimilarity(a, i.data[j].Values)
		if err != nil {
			return nil, err
		}
	}

	return scores, nil
}

func filterSearchResults(searchResults index.SearchResults, topK int) index.SearchResults {
	//sort by similarity score
	sort.Slice(searchResults, func(i, j int) bool {
		return (1 - searchResults[i].Score) < (1 - searchResults[j].Score)
	})

	maxTopK := topK
	if maxTopK > len(searchResults) {
		maxTopK = len(searchResults)
	}

	return searchResults[:maxTopK]
}
