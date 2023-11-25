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

var _ index.VectorDB = &DB{}

type data struct {
	ID       string     `json:"id"`
	Metadata types.Meta `json:"metadata"`
	Values   []float64  `json:"values"`
}

// DB is a simple in-memory vector database
// that stores the data in a json file only
// if the persist option is enabled.
type DB struct {
	data   []data
	dbPath string
}

type FilterFn func([]index.SearchResult) []index.SearchResult

func New() *DB {
	index := &DB{
		data: []data{},
	}

	return index
}

func (d *DB) WithPersist(dbPath string) *DB {
	d.dbPath = dbPath
	return d
}

func (d *DB) save() error {
	if d.dbPath == "" {
		return nil
	}

	jsonContent, err := json.Marshal(d.data)
	if err != nil {
		return err
	}

	return os.WriteFile(d.dbPath, jsonContent, 0600)
}

func (d *DB) load() error {
	if d.dbPath == "" {
		return nil
	}

	if len(d.data) > 0 {
		return nil
	}

	if _, err := os.Stat(d.dbPath); os.IsNotExist(err) {
		return d.save()
	}

	content, err := os.ReadFile(d.dbPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(content, &d.data)
}

func (d *DB) IsEmpty(_ context.Context) (bool, error) {
	err := d.load()
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return len(d.data) == 0, nil
}

func (d *DB) Insert(ctx context.Context, datas []index.Data) error {
	_ = ctx
	err := d.load()
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

	d.data = append(d.data, records...)

	return d.save()
}

func (d *DB) Search(ctx context.Context, values []float64, options *option.Options) (index.SearchResults, error) {
	err := d.load()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return d.similaritySearch(ctx, values, options)
}

func (d *DB) Drop(ctx context.Context) error {
	_ = ctx
	d.data = []data{}
	return d.save()
}

func (d *DB) Delete(ctx context.Context, ids []string) error {
	_ = ctx
	err := d.load()
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	var newRecords []data
	for _, record := range d.data {
		found := false
		for _, id := range ids {
			if record.ID == id {
				found = true
				break
			}
		}
		if !found {
			newRecords = append(newRecords, record)
		}
	}

	d.data = newRecords

	return d.save()
}

func (d *DB) similaritySearch(
	_ context.Context,
	embedding embedder.Embedding,
	opts *option.Options,
) (index.SearchResults, error) {
	if opts == nil {
		opts = index.GetDefaultOptions()
	}

	scores, err := d.cosineSimilarityBatch(embedding)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	searchResults := make([]index.SearchResult, len(scores))

	for j, score := range scores {
		searchResults[j] = index.SearchResult{
			Data: index.Data{
				ID:       d.data[j].ID,
				Values:   d.data[j].Values,
				Metadata: d.data[j].Metadata,
			},
			Score: score,
		}
	}

	if opts.Filter != nil {
		searchResults = opts.Filter.(FilterFn)(searchResults)
	}

	return filterSearchResults(searchResults, opts.TopK), nil
}

func (d *DB) cosineSimilarity(a []float64, b []float64) (cosine float64, err error) {
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

func (d *DB) cosineSimilarityBatch(a embedder.Embedding) ([]float64, error) {
	var err error
	scores := make([]float64, len(d.data))

	for j := range d.data {
		scores[j], err = d.cosineSimilarity(a, d.data[j].Values)
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
