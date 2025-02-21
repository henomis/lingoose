package pinecone

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	pineconego "github.com/henomis/pinecone-go/v2"
	pineconegorequest "github.com/henomis/pinecone-go/v2/request"
	pineconegoresponse "github.com/henomis/pinecone-go/v2/response"
	"github.com/rsest/lingoose/index"
	"github.com/rsest/lingoose/index/option"
)

var _ index.VectorDB = &DB{}

type DB struct {
	pineconeClient *pineconego.PineconeGo
	indexName      string
	namespace      string
	indexHost      *string

	createIndexOptions *CreateIndexOptions
}

type CreateIndexOptions struct {
	Dimension  int
	Metric     string
	Serverless *Serverless
	Pod        *Pod
}

type Serverless struct {
	Cloud  ServerlessCloud
	Region string
}

type ServerlessCloud string

const (
	ServerlessCloudAWS   ServerlessCloud = "aws"
	ServerlessCloudGCP   ServerlessCloud = "gcp"
	ServerlessCloudAzure ServerlessCloud = "azure"
)

type Pod struct {
	Environment      string
	Replicas         *int
	PodType          string
	Pods             *int
	Shards           *int
	MetadataConfig   *MetadataConfig
	SourceCollection *string
}

type MetadataConfig struct {
	Indexed []string
}

type Options struct {
	IndexName          string
	Namespace          string
	CreateIndexOptions *CreateIndexOptions
}

func New(options Options) *DB {
	apiKey := os.Getenv("PINECONE_API_KEY")

	pineconeClient := pineconego.New(apiKey)

	return &DB{
		pineconeClient:     pineconeClient,
		indexName:          options.IndexName,
		namespace:          options.Namespace,
		createIndexOptions: options.CreateIndexOptions,
	}
}

func (d *DB) WithAPIKey(apiKey string) *DB {
	d.pineconeClient = pineconego.New(apiKey)
	return d
}

func (d *DB) getIndexHost(ctx context.Context) error {
	if d.indexHost != nil {
		return nil
	}

	resp := &pineconegoresponse.IndexDescribe{}

	err := d.pineconeClient.IndexDescribe(ctx, &pineconegorequest.IndexDescribe{
		IndexName: d.indexName,
	}, resp)
	if err != nil {
		return err
	}

	resp.Host = "https://" + resp.Host
	d.indexHost = &resp.Host

	return nil
}

func (d *DB) IsEmpty(ctx context.Context) (bool, error) {
	err := d.createIndexIfRequired(ctx)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	err = d.getIndexHost(ctx)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	req := &pineconegorequest.VectorDescribeIndexStats{
		IndexHost: *d.indexHost,
	}
	res := &pineconegoresponse.VectorDescribeIndexStats{}

	err = d.pineconeClient.VectorDescribeIndexStats(ctx, req, res)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	namespace, ok := res.Namespaces[d.namespace]
	if !ok {
		return true, nil
	}

	if namespace.VectorCount == nil {
		return false, fmt.Errorf("%w: failed to get total index size", index.ErrInternal)
	}

	return *namespace.VectorCount == 0, nil
}

func (d *DB) Search(ctx context.Context, values []float64, options *option.Options) (index.SearchResults, error) {
	if options.Filter == nil {
		options.Filter = pineconegorequest.Filter{}
	}

	matches, err := d.similaritySearch(ctx, values, options)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return buildSearchResultsFromPineconeMatches(matches), nil
}

func (d *DB) Drop(ctx context.Context) error {
	err := d.pineconeClient.IndexDelete(ctx, &pineconegorequest.IndexDelete{
		IndexName: d.indexName,
	}, &pineconegoresponse.IndexDelete{})
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return nil
}

func (d *DB) Delete(ctx context.Context, ids []string) error {
	err := d.getIndexHost(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	deleteErr := d.pineconeClient.VectorDelete(ctx, &pineconegorequest.VectorDelete{
		IndexHost: *d.indexHost,
		IDs:       ids,
	}, &pineconegoresponse.VectorDelete{})
	if deleteErr != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, deleteErr)
	}

	return nil
}

func (d *DB) similaritySearch(
	ctx context.Context,
	values []float64,
	opts *option.Options,
) ([]pineconegoresponse.QueryMatch, error) {
	if opts == nil {
		opts = index.GetDefaultOptions()
	}

	if opts.Filter == nil {
		opts.Filter = pineconegorequest.Filter{}
	}

	err := d.getIndexHost(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	includeMetadata := true
	includeValues := true
	res := &pineconegoresponse.VectorQuery{}
	err = d.pineconeClient.VectorQuery(
		ctx,
		&pineconegorequest.VectorQuery{
			IndexHost:       *d.indexHost,
			TopK:            int32(opts.TopK),
			Vector:          values,
			IncludeMetadata: &includeMetadata,
			IncludeValues:   &includeValues,
			Namespace:       &d.namespace,
			Filter:          opts.Filter.(pineconegorequest.Filter),
		},
		res,
	)
	if err != nil {
		return nil, err
	}

	return res.Matches, nil
}

func (d *DB) createIndexIfRequired(ctx context.Context) error {
	if d.createIndexOptions == nil {
		return nil
	}

	resp := &pineconegoresponse.IndexList{}
	err := d.pineconeClient.IndexList(ctx, &pineconegorequest.IndexList{}, resp)
	if err != nil {
		return err
	}

	for _, index := range resp.Indexes {
		if index.Name == d.indexName {
			return nil
		}
	}

	metric := pineconegorequest.Metric(d.createIndexOptions.Metric)

	req := &pineconegorequest.IndexCreate{
		Name:      d.indexName,
		Dimension: d.createIndexOptions.Dimension,
		Metric:    &metric,
	}

	if d.createIndexOptions.Serverless != nil {
		req.Spec = pineconegorequest.Spec{
			Serverless: &pineconegorequest.ServerlessSpec{
				Cloud:  pineconegorequest.ServerlessSpecCloud(d.createIndexOptions.Serverless.Cloud),
				Region: d.createIndexOptions.Serverless.Region,
			},
		}
	} else if d.createIndexOptions.Pod != nil {
		req.Spec = pineconegorequest.Spec{
			Pod: &pineconegorequest.PodSpec{
				Environment:      d.createIndexOptions.Pod.Environment,
				Replicas:         d.createIndexOptions.Pod.Replicas,
				PodType:          d.createIndexOptions.Pod.PodType,
				Pods:             d.createIndexOptions.Pod.Pods,
				Shards:           d.createIndexOptions.Pod.Shards,
				MetadataConfig:   (*pineconegorequest.MetadataConfig)(d.createIndexOptions.Pod.MetadataConfig),
				SourceCollection: d.createIndexOptions.Pod.SourceCollection,
			},
		}
	}

	err = d.pineconeClient.IndexCreate(ctx, req, &pineconegoresponse.IndexCreate{})
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

			describe := &pineconegoresponse.IndexDescribe{}
			err = d.pineconeClient.IndexDescribe(ctx, &pineconegorequest.IndexDescribe{IndexName: d.indexName}, describe)
			if err != nil {
				return err
			}

			if describe.Status.Ready {
				return nil
			}

			time.Sleep(1 * time.Second)
		}
	}
}

func (d *DB) Insert(ctx context.Context, datas []index.Data) error {
	err := d.getIndexHost(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	vectors := []pineconegorequest.Vector{}
	for _, data := range datas {
		if data.ID == "" {
			id, errUUID := uuid.NewUUID()
			if errUUID != nil {
				return errUUID
			}
			data.ID = id.String()
		}

		vector := pineconegorequest.Vector{
			ID:       data.ID,
			Values:   data.Values,
			Metadata: data.Metadata,
		}
		vectors = append(vectors, vector)
	}

	req := &pineconegorequest.VectorUpsert{
		IndexHost: *d.indexHost,
		Vectors:   vectors,
		Namespace: d.namespace,
	}
	res := &pineconegoresponse.VectorUpsert{}

	err = d.pineconeClient.VectorUpsert(ctx, req, res)
	if err != nil {
		return err
	}

	if res.UpsertedCount == nil || res.UpsertedCount != nil && *res.UpsertedCount != int64(len(vectors)) {
		return fmt.Errorf("error upserting embeddings")
	}

	return nil
}

func buildSearchResultsFromPineconeMatches(
	matches []pineconegoresponse.QueryMatch,
) index.SearchResults {
	searchResults := make([]index.SearchResult, len(matches))

	for i, match := range matches {
		metadata := index.DeepCopyMetadata(match.Metadata)

		id := ""
		if match.ID != nil {
			id = *match.ID
		}

		score := float64(0)
		if match.Score != nil {
			score = *match.Score
		}

		searchResults[i] = index.SearchResult{
			Data: index.Data{
				ID:       id,
				Metadata: metadata,
				Values:   match.Values,
			},
			Score: score,
		}
	}

	return searchResults
}
