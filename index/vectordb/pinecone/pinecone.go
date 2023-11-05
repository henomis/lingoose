package pinecone

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/option"
	pineconego "github.com/henomis/pinecone-go"
	pineconegorequest "github.com/henomis/pinecone-go/request"
	pineconegoresponse "github.com/henomis/pinecone-go/response"
)

type DB struct {
	pineconeClient *pineconego.PineconeGo
	indexName      string
	projectID      *string
	namespace      string

	createIndexOptions *CreateIndexOptions
}

type CreateIndexOptions struct {
	Dimension int
	Replicas  int
	Metric    string
	PodType   string
}

type Options struct {
	IndexName          string
	Namespace          string
	CreateIndexOptions *CreateIndexOptions
}

func New(options Options) *DB {
	apiKey := os.Getenv("PINECONE_API_KEY")
	environment := os.Getenv("PINECONE_ENVIRONMENT")

	pineconeClient := pineconego.New(environment, apiKey)

	return &DB{
		pineconeClient:     pineconeClient,
		indexName:          options.IndexName,
		namespace:          options.Namespace,
		createIndexOptions: options.CreateIndexOptions,
	}
}

func (d *DB) WithAPIKeyAndEnvironment(apiKey, environment string) *DB {
	d.pineconeClient = pineconego.New(environment, apiKey)
	return d
}

func (d *DB) IsEmpty(ctx context.Context) (bool, error) {
	err := d.createIndexIfRequired(ctx)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	err = d.getProjectID(ctx)
	if err != nil {
		return true, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	req := &pineconegorequest.VectorDescribeIndexStats{
		IndexName: d.indexName,
		ProjectID: *d.projectID,
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
		options.Filter = map[string]string{}
	}

	matches, err := d.similaritySearch(ctx, values, options)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	return buildSearchResultsFromPineconeMatches(matches), nil
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
		opts.Filter = map[string]string{}
	}

	err := d.getProjectID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", index.ErrInternal, err)
	}

	includeMetadata := true
	includeValues := true
	res := &pineconegoresponse.VectorQuery{}
	err = d.pineconeClient.VectorQuery(
		ctx,
		&pineconegorequest.VectorQuery{
			IndexName:       d.indexName,
			ProjectID:       *d.projectID,
			TopK:            int32(opts.TopK),
			Vector:          values,
			IncludeMetadata: &includeMetadata,
			IncludeValues:   &includeValues,
			Namespace:       &d.namespace,
			Filter:          opts.Filter.(map[string]string),
		},
		res,
	)
	if err != nil {
		return nil, err
	}

	return res.Matches, nil
}

func (d *DB) getProjectID(ctx context.Context) error {
	if d.projectID != nil {
		return nil
	}

	whoamiResp := &pineconegoresponse.Whoami{}

	err := d.pineconeClient.Whoami(ctx, &pineconegorequest.Whoami{}, whoamiResp)
	if err != nil {
		return err
	}

	d.projectID = &whoamiResp.ProjectID

	return nil
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
		if index == d.indexName {
			return nil
		}
	}

	metric := pineconegorequest.Metric(d.createIndexOptions.Metric)

	req := &pineconegorequest.IndexCreate{
		Name:      d.indexName,
		Dimension: d.createIndexOptions.Dimension,
		Replicas:  &d.createIndexOptions.Replicas,
		Metric:    &metric,
		PodType:   &d.createIndexOptions.PodType,
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
	err := d.getProjectID(ctx)
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
		IndexName: d.indexName,
		ProjectID: *d.projectID,
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
