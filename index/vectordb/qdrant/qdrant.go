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

type DB struct {
	qdrantClient   *qdrantgo.Client
	collectionName string

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
	CreateCollection *CreateCollectionOptions
}

func New(options Options) *DB {
	apiKey := os.Getenv("QDRANT_API_KEY")
	endpoint := os.Getenv("QDRANT_ENDPOINT")

	qdrantClient := qdrantgo.New(endpoint, apiKey)

	return &DB{
		qdrantClient:     qdrantClient,
		collectionName:   options.CollectionName,
		createCollection: options.CreateCollection,
	}
}

func (d *DB) WithAPIKeyAndEdpoint(apiKey, endpoint string) *DB {
	d.qdrantClient = qdrantgo.New(endpoint, apiKey)
	return d
}

func (d *DB) IsEmpty(ctx context.Context) (bool, error) {
	err := d.createCollectionIfRequired(ctx)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	res := &qdrantresponse.CollectionCollectInfo{}
	err = d.qdrantClient.CollectionCollectInfo(
		ctx,
		&qdrantrequest.CollectionCollectInfo{
			CollectionName: d.collectionName,
		},
		res,
	)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return res.Result.VectorsCount == 0, nil
}

func (d *DB) Insert(ctx context.Context, datas []index.Data) error {
	err := d.createCollectionIfRequired(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	var points []qdrantrequest.Point
	for _, data := range datas {
		if data.ID == "" {
			id, errUUID := uuid.NewUUID()
			if errUUID != nil {
				return errUUID
			}
			data.ID = id.String()
		}

		point := qdrantrequest.Point{
			ID:      data.ID,
			Vector:  data.Values,
			Payload: data.Metadata,
		}
		points = append(points, point)
	}

	wait := true
	req := &qdrantrequest.PointUpsert{
		Wait:           &wait,
		CollectionName: d.collectionName,
		Points:         points,
	}
	res := &qdrantresponse.PointUpsert{}

	return d.qdrantClient.PointUpsert(ctx, req, res)
}

func (d *DB) Search(ctx context.Context, values []float64, options *option.Options) (index.SearchResults, error) {
	matches, err := d.similaritySearch(ctx, values, options)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return buildSearchResultsFromQdrantMatches(matches), nil
}

func (d *DB) similaritySearch(
	ctx context.Context,
	values []float64,
	opts *option.Options,
) ([]qdrantresponse.PointSearchResult, error) {
	if opts.Filter == nil {
		opts.Filter = qdrantrequest.Filter{}
	}

	includeMetadata := true
	includeValues := true
	res := &qdrantresponse.PointSearch{}
	err := d.qdrantClient.PointSearch(
		ctx,
		&qdrantrequest.PointSearch{
			CollectionName: d.collectionName,
			Limit:          opts.TopK,
			Vector:         values,
			WithPayload:    &includeMetadata,
			WithVector:     &includeValues,
			Filter:         opts.Filter.(qdrantrequest.Filter),
		},
		res,
	)
	if err != nil {
		return nil, err
	}

	return res.Result, nil
}

func (d *DB) createCollectionIfRequired(ctx context.Context) error {
	if d.createCollection == nil {
		return nil
	}

	resp := &qdrantresponse.CollectionList{}
	err := d.qdrantClient.CollectionList(ctx, &qdrantrequest.CollectionList{}, resp)
	if err != nil {
		return err
	}

	for _, collection := range resp.Result.Collections {
		if collection.Name == d.collectionName {
			return nil
		}
	}

	req := &qdrantrequest.CollectionCreate{
		CollectionName: d.collectionName,
		Vectors: qdrantrequest.VectorsParams{
			Size:     d.createCollection.Dimension,
			Distance: qdrantrequest.Distance(d.createCollection.Distance),
			OnDisk:   &d.createCollection.OnDisk,
		},
	}

	err = d.qdrantClient.CollectionCreate(ctx, req, &qdrantresponse.CollectionCreate{})
	if err != nil {
		return err
	}

	return nil
}

func buildSearchResultsFromQdrantMatches(
	matches []qdrantresponse.PointSearchResult,
) index.SearchResults {
	searchResults := make([]index.SearchResult, len(matches))

	for i, match := range matches {
		metadata := index.DeepCopyMetadata(match.Payload)

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
