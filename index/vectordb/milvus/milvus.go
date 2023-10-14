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
	IncludeContent   bool
	IncludeValues    bool
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

func (q *DB) WithCredentialsAndEndpoint(username, password, endpoint string) *DB {
	q.milvusClient = milvusgo.New(endpoint, username, password)
	return q
}

// func (q *Index) LoadFromDocuments(ctx context.Context, documents []document.Document) error {
// 	err := q.createCollectionIfRequired(ctx)
// 	if err != nil {
// 		return fmt.Errorf("%w: %w", index.ErrInternal, err)
// 	}

// 	err = q.batchUpsert(ctx, documents)
// 	if err != nil {
// 		return fmt.Errorf("%w: %w", index.ErrInternal, err)
// 	}
// 	return nil
// }

func (q *DB) IsEmpty(ctx context.Context) (bool, error) {
	err := q.createCollectionIfRequired(ctx)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	vector := make([]float64, q.createCollection.Dimension)
	res := &milvusgoresponse.VectorSearch{}
	err = q.milvusClient.VectorSearch(
		ctx,
		&milvusgorequest.VectorSearch{
			CollectionName: q.collectionName,
			Vector:         vector,
		},
		res,
	)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return len(res.Data) == 0, nil
}

// func (q *Index) Add(ctx context.Context, item *index.Data) error {
// 	err := q.createCollectionIfRequired(ctx)
// 	if err != nil {
// 		return fmt.Errorf("%w: %w", index.ErrInternal, err)
// 	}

// 	vectorData := make(milvusrequest.VectorData, 0)

// 	for k, v := range item.Metadata {
// 		vectorData[k] = v
// 	}
// 	vectorData[milvusrequest.DefaultVectorField] = item.Values

// 	return q.pointUpsert(ctx,
// 		[]milvusrequest.VectorData{
// 			vectorData,
// 		},
// 	)
// }

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
	if opts.Filter == nil {
		opts.Filter = ""
	}

	filter := opts.Filter.(string)

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

// func (q *Index) query(
// 	ctx context.Context,
// 	query string,
// 	opts *option.Options,
// ) ([]milvusresponse.VectorData, error) {
// 	embeddings, err := q.embedder.Embed(ctx, []string{query})
// 	if err != nil {
// 		return nil, err
// 	}

// 	return q.similaritySearch(ctx, embeddings[0], opts)
// }

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

// func (q *Index) batchUpsert(ctx context.Context, documents []document.Document) error {
// 	for i := 0; i < len(documents); i += q.batchUpsertSize {
// 		batchEnd := i + q.batchUpsertSize
// 		if batchEnd > len(documents) {
// 			batchEnd = len(documents)
// 		}

// 		texts := []string{}
// 		for _, document := range documents[i:batchEnd] {
// 			texts = append(texts, document.Content)
// 		}

// 		embeddings, err := q.embedder.Embed(ctx, texts)
// 		if err != nil {
// 			return err
// 		}

// 		points, err := buildMilvusPointsFromEmbeddingsAndDocuments(embeddings, documents, i, q.includeContent)
// 		if err != nil {
// 			return err
// 		}

// 		err = q.pointUpsert(ctx, points)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// TODO: rename vectorInsert
func (q *DB) Insert(ctx context.Context, datas []index.Data) error {

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
		CollectionName: q.collectionName,
		Data:           vectors,
	}
	res := &milvusgoresponse.VectorInsert{}

	err := q.milvusClient.VectorInsert(ctx, req, res)
	if err != nil {
		return err
	}

	return nil
}

// func buildMilvusPointsFromEmbeddingsAndDocuments(
// 	embeddings []embedder.Embedding,
// 	documents []document.Document,
// 	startIndex int,
// 	includeContent bool,
// ) ([]milvusrequest.VectorData, error) {
// 	var vectors []milvusrequest.VectorData

// 	for i, embedding := range embeddings {
// 		metadata := index.DeepCopyMetadata(documents[startIndex+i].Metadata)

// 		// inject document content into vector metadata
// 		if includeContent {
// 			metadata[index.DefaultKeyContent] = documents[startIndex+i].Content
// 		}

// 		vectorData := make(milvusrequest.VectorData, 0)
// 		for k, v := range metadata {
// 			vectorData[k] = v
// 		}
// 		vectorData[milvusrequest.DefaultVectorField] = embedding

// 		vectors = append(vectors, vectorData)

// 		// inject vector ID into document metadata
// 		// documents[startIndex+i].Metadata[index.DefaultKeyID] = vectorID.String()
// 	}

// 	return vectors, nil
// }

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
