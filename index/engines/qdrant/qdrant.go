package qdrant

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/option"
	qdrantgo "github.com/henomis/qdrant-go"
	qdrantrequest "github.com/henomis/qdrant-go/request"
	qdrantresponse "github.com/henomis/qdrant-go/response"
)

const (
	defaultTopK = 10
)

type IndexEngine struct {
	qdrantClient   *qdrantgo.Client
	collectionName string
	includeContent bool
	includeValues  bool

	createCollection *CreateCollectionOptions
}

type Distance string

const (
	DistanceCosine    Distance = Distance(qdrantrequest.DistanceCosine)
	DistanceEuclidean Distance = Distance(qdrantrequest.DistanceEuclidean)
	DistanceDot       Distance = Distance(qdrantrequest.DistanceDot)
)

type CreateCollectionOptions struct {
	Dimension uint64
	Distance  Distance
	OnDisk    bool
}

type Options struct {
	CollectionName   string
	IncludeContent   bool
	IncludeValues    bool
	BatchUpsertSize  *int
	CreateCollection *CreateCollectionOptions
}

func New(options Options) *IndexEngine {
	apiKey := os.Getenv("QDRANT_API_KEY")
	endpoint := os.Getenv("QDRANT_ENDPOINT")

	qdrantClient := qdrantgo.New(endpoint, apiKey)

	return &IndexEngine{
		qdrantClient:     qdrantClient,
		collectionName:   options.CollectionName,
		includeContent:   options.IncludeContent,
		includeValues:    options.IncludeValues,
		createCollection: options.CreateCollection,
	}
}

func (i *IndexEngine) WithAPIKeyAndEdpoint(apiKey, endpoint string) *IndexEngine {
	i.qdrantClient = qdrantgo.New(endpoint, apiKey)
	return i
}

func (q *IndexEngine) IsEmpty(ctx context.Context) (bool, error) {
	err := q.createCollectionIfRequired(ctx)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	res := &qdrantresponse.CollectionCollectInfo{}
	err = q.qdrantClient.CollectionCollectInfo(
		ctx,
		&qdrantrequest.CollectionCollectInfo{
			CollectionName: q.collectionName,
		},
		res,
	)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return res.Result.VectorsCount == 0, nil
}

func (i *IndexEngine) Insert(ctx context.Context, data []index.Data) error {
	err := i.createCollectionIfRequired(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	var points []qdrantrequest.Point
	for _, d := range data {
		if d.ID == "" {
			id, errUUID := uuid.NewUUID()
			if errUUID != nil {
				return errUUID
			}
			d.ID = id.String()
		}

		point := qdrantrequest.Point{
			ID:      d.ID,
			Vector:  d.Values,
			Payload: d.Metadata,
		}
		points = append(points, point)
	}

	wait := true
	req := &qdrantrequest.PointUpsert{
		Wait:           &wait,
		CollectionName: i.collectionName,
		Points:         points,
	}
	res := &qdrantresponse.PointUpsert{}

	return i.qdrantClient.PointUpsert(ctx, req, res)
}

func (i *IndexEngine) Search(ctx context.Context, values []float64, options *option.Options) (index.SearchResults, error) {
	matches, err := i.similaritySearch(ctx, values, options)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return buildSearchResultsFromQdrantMatches(matches, i.includeContent), nil
}

func (q *IndexEngine) similaritySearch(
	ctx context.Context,
	values []float64,
	opts *option.Options,
) ([]qdrantresponse.PointSearchResult, error) {
	if opts.Filter == nil {
		opts.Filter = qdrantrequest.Filter{}
	}

	includeMetadata := true
	res := &qdrantresponse.PointSearch{}
	err := q.qdrantClient.PointSearch(
		ctx,
		&qdrantrequest.PointSearch{
			CollectionName: q.collectionName,
			Limit:          opts.TopK,
			Vector:         values,
			WithPayload:    &includeMetadata,
			WithVector:     &q.includeValues,
			Filter:         opts.Filter.(qdrantrequest.Filter),
		},
		res,
	)
	if err != nil {
		return nil, err
	}

	return res.Result, nil
}

func (q *IndexEngine) createCollectionIfRequired(ctx context.Context) error {
	if q.createCollection == nil {
		return nil
	}

	resp := &qdrantresponse.CollectionList{}
	err := q.qdrantClient.CollectionList(ctx, &qdrantrequest.CollectionList{}, resp)
	if err != nil {
		return err
	}

	for _, collection := range resp.Result.Collections {
		if collection.Name == q.collectionName {
			return nil
		}
	}

	req := &qdrantrequest.CollectionCreate{
		CollectionName: q.collectionName,
		Vectors: qdrantrequest.VectorsParams{
			Size:     q.createCollection.Dimension,
			Distance: qdrantrequest.Distance(q.createCollection.Distance),
			OnDisk:   &q.createCollection.OnDisk,
		},
	}

	err = q.qdrantClient.CollectionCreate(ctx, req, &qdrantresponse.CollectionCreate{})
	if err != nil {
		return err
	}

	return nil
}

func buildSearchResultsFromQdrantMatches(
	matches []qdrantresponse.PointSearchResult,
	includeContent bool,
) index.SearchResults {
	searchResults := make([]index.SearchResult, len(matches))

	for i, match := range matches {
		metadata := index.DeepCopyMetadata(match.Payload)
		if !includeContent {
			delete(metadata, index.DefaultKeyContent)
		}

		searchResults[i] = index.SearchResult{
			Data: index.Data{
				ID:       match.ID,
				Metadata: metadata,
				Values:   match.Vector,
			},
			Score: match.Score,
		}
	}

	return searchResults
}
