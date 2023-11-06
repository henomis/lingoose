package milvus

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/option"
	"github.com/henomis/lingoose/types"
	milvusgo "github.com/henomis/milvus-go"
	milvusgorequest "github.com/henomis/milvus-go/request"
	milvusgoresponse "github.com/henomis/milvus-go/response"
)

type DB struct {
	milvusClient   *milvusgo.Client
	databaseName   *string
	collectionName string

	createCollection *CreateCollectionOptions
}

type Metric string

const (
	DistanceL2             Metric = Metric(milvusgorequest.MetricL2)
	DistanceIP             Metric = Metric(milvusgorequest.MetricIP)
	DistanceHamming        Metric = Metric(milvusgorequest.MetricHamming)
	DistanceJaccard        Metric = Metric(milvusgorequest.MetricJaccard)
	DistanceTanimoto       Metric = Metric(milvusgorequest.MetricTanimoto)
	DistanceSubStructure   Metric = Metric(milvusgorequest.MetricSubstructure)
	DistanceSuperStructure Metric = Metric(milvusgorequest.MetricSuperstructure)
)

type CreateCollectionOptions struct {
	Dimension uint64
	Metric    Metric
}

type Options struct {
	DatabaseName     *string
	CollectionName   string
	BatchUpsertSize  *int
	CreateCollection *CreateCollectionOptions
}

func New(options Options) *DB {
	username := os.Getenv("MILVUS_USERNAME")
	password := os.Getenv("MILVUS_PASSWORD")
	endpoint := os.Getenv("MILVUS_ENDPOINT")

	milvusClient := milvusgo.New(endpoint, username, password)

	return &DB{
		milvusClient:     milvusClient,
		databaseName:     options.DatabaseName,
		collectionName:   options.CollectionName,
		createCollection: options.CreateCollection,
	}
}

func (d *DB) WithCredentialsAndEndpoint(username, password, endpoint string) *DB {
	d.milvusClient = milvusgo.New(endpoint, username, password)
	return d
}

func (d *DB) IsEmpty(ctx context.Context) (bool, error) {
	err := d.createCollectionIfRequired(ctx)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	vector := make([]float64, d.createCollection.Dimension)
	res := &milvusgoresponse.VectorSearch{}
	err = d.milvusClient.VectorSearch(
		ctx,
		&milvusgorequest.VectorSearch{
			CollectionName: d.collectionName,
			Vector:         vector,
		},
		res,
	)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return len(res.Data) == 0, nil
}

func (d *DB) Search(ctx context.Context, values []float64, options *option.Options) (index.SearchResults, error) {
	matches, err := d.similaritySearch(ctx, values, options)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return buildSearchResultsFromMilvusMatches(matches), nil
}

func (d *DB) similaritySearch(
	ctx context.Context,
	values []float64,
	opts *option.Options,
) ([]milvusgoresponse.VectorData, error) {
	if opts == nil {
		opts = index.GetDefaultOptions()
	}

	if opts.Filter == nil {
		opts.Filter = ""
	}

	filter, ok := opts.Filter.(string)
	if !ok {
		return nil, fmt.Errorf("invalid filter")
	}

	limit := uint64(opts.TopK)

	outputFields := []string{
		index.DefaultKeyContent,
		milvusgorequest.DefaultPrimaryField,
		milvusgorequest.DefaultDistanceField,
		milvusgorequest.DefaultVectorField,
	}

	res := &milvusgoresponse.VectorSearch{}
	req := &milvusgorequest.VectorSearch{
		CollectionName: d.collectionName,
		Limit:          &limit,
		Vector:         values,
		OutputFields:   outputFields,
	}

	if filter != "" {
		req.Filter = &filter
	}

	err := d.milvusClient.VectorSearch(
		ctx,
		req,
		res,
	)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}

func (d *DB) createCollectionIfRequired(ctx context.Context) error {
	if d.createCollection == nil {
		return nil
	}

	resp := &milvusgoresponse.CollectionList{}
	err := d.milvusClient.CollectionList(ctx, &milvusgorequest.CollectionList{}, resp)
	if err != nil {
		return err
	}

	for _, collection := range resp.Data {
		if collection == d.collectionName {
			return nil
		}
	}

	req := &milvusgorequest.CollectionCreate{
		CollectionName: d.collectionName,
		DBName:         d.databaseName,
		Dimension:      d.createCollection.Dimension,
	}

	metric := milvusgorequest.Metric(d.createCollection.Metric)
	req.MetricType = &metric

	err = d.milvusClient.CollectionCreate(ctx, req, &milvusgoresponse.CollectionCreate{})
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) Insert(ctx context.Context, datas []index.Data) error {
	var vectors []milvusgorequest.VectorData
	for _, data := range datas {
		// uncomment this as soon as milvus rest supports ids
		// if data.ID == "" {
		// 	id, errUUID := uuid.NewUUID()
		// 	if errUUID != nil {
		// 		return errUUID
		// 	}
		// 	data.ID = id.String()
		// }

		vectorData := make(milvusgorequest.VectorData, 0)
		for k, v := range data.Metadata {
			vectorData[k] = v
		}
		vectorData[milvusgorequest.DefaultVectorField] = data.Values
		vectors = append(vectors, vectorData)
	}

	req := &milvusgorequest.VectorInsert{
		CollectionName: d.collectionName,
		Data:           vectors,
	}
	res := &milvusgoresponse.VectorInsert{}

	err := d.milvusClient.VectorInsert(ctx, req, res)
	if err != nil {
		return err
	}

	return nil
}

func buildSearchResultsFromMilvusMatches(
	matches []milvusgoresponse.VectorData,
) index.SearchResults {
	searchResults := make([]index.SearchResult, len(matches))

	for i, match := range matches {
		metadata := make(types.Meta, 0)
		for k, v := range match {
			if k == milvusgorequest.DefaultVectorField || k == milvusgorequest.DefaultPrimaryField {
				continue
			}

			metadata[k] = v
		}

		distance, err := match[milvusgorequest.DefaultDistanceField].(json.Number).Float64()
		if err != nil {
			return nil
		}

		searchResults[i] = index.SearchResult{
			Data: index.Data{
				ID:       fmt.Sprintf("%d", match.ID()),
				Metadata: metadata,
				Values:   match.Vector(),
			},
			Score: distance,
		}
	}

	return searchResults
}
