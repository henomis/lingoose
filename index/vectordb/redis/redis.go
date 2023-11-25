package redis

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/option"

	"github.com/RediSearch/redisearch-go/v2/redisearch"
)

var _ index.VectorDB = &DB{}

const (
	errUnknownIndexName = "Unknown index name"
)

type DB struct {
	redisearchClient *redisearch.Client
	createIndex      *CreateIndexOptions
}

type Distance string

const (
	DistanceCosine       Distance = "COSINE"
	DistanceEuclidean    Distance = "L2"
	DistanceInnerProduct Distance = "IP"

	defaultVectorFieldName      = "vec"
	defaultVectorScoreFieldName = "__vec_score"
)

type CreateIndexOptions struct {
	Dimension uint64
	Distance  Distance
}

type Options struct {
	RedisearchClient *redisearch.Client
	CreateIndex      *CreateIndexOptions
}

func New(options Options) *DB {
	return &DB{
		redisearchClient: options.RedisearchClient,
		createIndex:      options.CreateIndex,
	}
}

func (d *DB) IsEmpty(ctx context.Context) (bool, error) {
	err := d.createIndexIfRequired(ctx)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	indexInfo, err := d.redisearchClient.Info()
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return indexInfo.DocCount == 0, nil
}

func (d *DB) Insert(ctx context.Context, datas []index.Data) error {
	err := d.createIndexIfRequired(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	var documents []redisearch.Document
	for _, data := range datas {
		if data.ID == "" {
			id, errUUID := uuid.NewUUID()
			if errUUID != nil {
				return errUUID
			}
			data.ID = id.String()
		}

		document := redisearch.NewDocument(data.ID, 1.0)

		for key, value := range data.Metadata {
			document.Set(key, value)
		}

		document.Set(defaultVectorFieldName, float64tobytes(data.Values))

		documents = append(documents, document)
	}

	if err = d.redisearchClient.Index(documents...); err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return nil
}

func (d *DB) Search(ctx context.Context, values []float64, options *option.Options) (index.SearchResults, error) {
	matches, err := d.similaritySearch(ctx, values, options)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return buildSearchResultsFromRedisDocuments(matches), nil
}

func (d *DB) Drop(ctx context.Context) error {
	err := d.redisearchClient.Drop()
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return nil
}

func (d *DB) Delete(ctx context.Context, ids []string) error {
	for _, id := range ids {
		err := d.redisearchClient.DeleteDocument(id)
		if err != nil {
			return fmt.Errorf("%w: %w", index.ErrInternal, err)
		}
	}

	return nil
}

func (d *DB) similaritySearch(
	_ context.Context,
	values []float64,
	opts *option.Options,
) ([]redisearch.Document, error) {
	if opts == nil {
		opts = index.GetDefaultOptions()
	}

	if opts.Filter == nil {
		opts.Filter = redisearch.Filter{}
	}

	docs, _, err := d.redisearchClient.Search(
		redisearch.NewQuery(fmt.Sprintf("(*)=>[KNN %d @vec $query_vector]", opts.TopK)).
			SetSortBy(defaultVectorScoreFieldName, true).
			SetFlags(redisearch.QueryWithPayloads).
			SetDialect(2).
			Limit(0, opts.TopK).
			AddParam("query_vector", float64tobytes(values)).
			AddFilter(opts.Filter.(redisearch.Filter)),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return docs, nil
}

func (d *DB) createIndexIfRequired(_ context.Context) error {
	if d.createIndex == nil {
		return nil
	}

	indexName := ""
	indexInfo, err := d.redisearchClient.Info()
	if err != nil && (err.Error() != errUnknownIndexName) {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	} else if err == nil {
		indexName = indexInfo.Name
	}

	indexes, err := d.redisearchClient.List()
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	if len(indexes) > 0 && len(indexName) > 0 {
		for _, index := range indexes {
			if index == indexInfo.Name {
				return nil
			}
		}
	}

	err = d.redisearchClient.CreateIndex(
		redisearch.NewSchema(redisearch.DefaultOptions).
			AddField(redisearch.NewVectorFieldOptions(
				defaultVectorFieldName,
				redisearch.VectorFieldOptions{
					Algorithm: redisearch.Flat,
					Attributes: map[string]interface{}{
						"TYPE":            "FLOAT32",
						"DIM":             d.createIndex.Dimension,
						"DISTANCE_METRIC": d.createIndex.Distance,
					}})),
	)
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return nil
}

func buildSearchResultsFromRedisDocuments(
	documents []redisearch.Document,
) index.SearchResults {
	searchResults := make([]index.SearchResult, len(documents))

	for i, match := range documents {
		metadata := index.DeepCopyMetadata(match.Properties)

		score := 0.0
		scoreField, fieldExists := match.Properties[defaultVectorScoreFieldName]
		if fieldExists {
			scoreAsString, ok := scoreField.(string)
			if ok {
				score, _ = strconv.ParseFloat(scoreAsString, 64)
				delete(metadata, defaultVectorScoreFieldName)
			}
		}

		values := []float64{}
		vectorField, fieldExists := match.Properties[defaultVectorFieldName]
		if fieldExists {
			vectorAsString, ok := vectorField.(string)
			if ok {
				values = bytestofloat64([]byte(vectorAsString))
				delete(metadata, defaultVectorFieldName)
			}
		}

		searchResults[i] = index.SearchResult{
			Data: index.Data{
				ID:       match.Id,
				Metadata: metadata,
				Values:   values,
			},
			Score: score,
		}
	}

	return searchResults
}

func float64to32(floats []float64) []float32 {
	floats32 := make([]float32, len(floats))
	for i, f := range floats {
		floats32[i] = float32(f)
	}
	return floats32
}

func float64tobytes(floats64 []float64) []byte {
	floats := float64to32(floats64)

	byteSlice := make([]byte, len(floats)*4)
	for i, f := range floats {
		bits := math.Float32bits(f)
		binary.LittleEndian.PutUint32(byteSlice[i*4:], bits)
	}
	return byteSlice
}
func bytestofloat64(byteSlice []byte) []float64 {
	floats := make([]float32, len(byteSlice)/4)
	for i := 0; i < len(byteSlice); i += 4 {
		bits := binary.LittleEndian.Uint32(byteSlice[i : i+4])
		floats[i/4] = math.Float32frombits(bits)
	}

	floats64 := make([]float64, len(floats))
	for i, f := range floats {
		floats64[i] = float64(f)
	}
	return floats64
}
