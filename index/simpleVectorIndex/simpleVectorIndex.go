package simplevectorindex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/option"
	"github.com/henomis/lingoose/types"
)

const (
	defaultBatchSize = 32
	defaultTopK      = 10
)

type data struct {
	ID       string     `json:"id"`
	Metadata types.Meta `json:"metadata"`
	Values   []float64  `json:"values"`
}

type Index struct {
	data       []data
	outputPath string
	name       string
	embedder   index.Embedder
}

type FilterFn func([]index.SearchResult) []index.SearchResult

func New(name string, outputPath string, embedder index.Embedder) *Index {
	simpleVectorIndex := &Index{
		data:       []data{},
		outputPath: outputPath,
		name:       name,
		embedder:   embedder,
	}

	return simpleVectorIndex
}

func (s *Index) LoadFromDocuments(ctx context.Context, documents []document.Document) error {
	err := s.load()
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	for i := 0; i < len(documents); i += defaultBatchSize {
		end := i + defaultBatchSize
		if end > len(documents) {
			end = len(documents)
		}

		texts := []string{}
		for _, document := range documents[i:end] {
			texts = append(texts, document.Content)
		}

		embeddings, errEmbed := s.embedder.Embed(ctx, texts)
		if errEmbed != nil {
			return fmt.Errorf("%w: %w", index.ErrInternal, errEmbed)
		}

		for j, document := range documents[i:end] {
			id, errUUID := uuid.NewUUID()
			if errUUID != nil {
				return errUUID
			}
			s.data = append(s.data, buildDataFromEmbeddingAndDocument(id.String(), embeddings[j], document))
		}
	}

	err = s.save()
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return nil
}

func buildDataFromEmbeddingAndDocument(
	id string,
	embedding embedder.Embedding,
	document document.Document,
) data {
	metadata := index.DeepCopyMetadata(document.Metadata)
	metadata[index.DefaultKeyContent] = document.Content
	return data{
		ID:       id,
		Values:   embedding,
		Metadata: metadata,
	}
}

func (s Index) save() error {
	jsonContent, err := json.Marshal(s.data)
	if err != nil {
		return err
	}

	return os.WriteFile(s.database(), jsonContent, 0600)
}

func (s *Index) load() error {
	if len(s.data) > 0 {
		return nil
	}

	if _, err := os.Stat(s.database()); os.IsNotExist(err) {
		return s.save()
	}

	content, err := os.ReadFile(s.database())
	if err != nil {
		return err
	}

	return json.Unmarshal(content, &s.data)
}

func (s *Index) database() string {
	return strings.Join([]string{s.outputPath, s.name + ".json"}, string(os.PathSeparator))
}

func (s *Index) IsEmpty() (bool, error) {
	err := s.load()
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return len(s.data) == 0, nil
}

func (s *Index) Add(ctx context.Context, item *index.Data) error {
	_ = ctx
	err := s.load()
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	if item.ID == "" {
		id, errUUID := uuid.NewUUID()
		if errUUID != nil {
			return errUUID
		}
		item.ID = id.String()
	}

	s.data = append(
		s.data,
		data{
			ID:       item.ID,
			Values:   item.Values,
			Metadata: item.Metadata,
		},
	)

	return s.save()
}

func (s *Index) Search(ctx context.Context, values []float64, opts ...option.Option) (index.SearchResults, error) {
	sviOptions := &option.Options{
		TopK: defaultTopK,
	}

	for _, opt := range opts {
		opt(sviOptions)
	}

	err := s.load()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return s.similaritySearch(ctx, values, sviOptions)
}

func (s *Index) Query(ctx context.Context, query string, opts ...option.Option) (index.SearchResults, error) {
	sviOptions := &option.Options{
		TopK: defaultTopK,
	}

	for _, opt := range opts {
		opt(sviOptions)
	}

	err := s.load()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	embeddings, err := s.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return s.similaritySearch(ctx, embeddings[0], sviOptions)
}

func (s *Index) similaritySearch(
	ctx context.Context,
	embedding embedder.Embedding,
	opts *option.Options,
) (index.SearchResults, error) {
	_ = ctx
	scores, err := s.cosineSimilarityBatch(embedding)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	searchResults := make([]index.SearchResult, len(scores))

	for i, score := range scores {
		searchResults[i] = index.SearchResult{
			Data: index.Data{
				ID:       s.data[i].ID,
				Values:   s.data[i].Values,
				Metadata: s.data[i].Metadata,
			},
			Score: score,
		}
	}

	if opts.Filter != nil {
		searchResults = opts.Filter.(FilterFn)(searchResults)
	}

	return filterSearchResults(searchResults, opts.TopK), nil
}

func (s *Index) cosineSimilarity(a []float64, b []float64) (cosine float64, err error) {
	count := 0
	length_a := len(a)
	length_b := len(b)
	if length_a > length_b {
		count = length_a
	} else {
		count = length_b
	}
	sumA := 0.0
	s1 := 0.0
	s2 := 0.0
	for k := 0; k < count; k++ {
		if k >= length_a {
			s2 += math.Pow(b[k], 2)
			continue
		}
		if k >= length_b {
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

func (s *Index) cosineSimilarityBatch(a embedder.Embedding) ([]float64, error) {
	var err error
	scores := make([]float64, len(s.data))

	for i := range s.data {
		scores[i], err = s.cosineSimilarity(a, s.data[i].Values)
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
