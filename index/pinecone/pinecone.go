package pinecone

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/option"
	pineconego "github.com/henomis/pinecone-go"
	pineconerequest "github.com/henomis/pinecone-go/request"
	pineconeresponse "github.com/henomis/pinecone-go/response"
)

const (
	defaultTopK            = 10
	defaultBatchUpsertSize = 32
)

type Index struct {
	pineconeClient  *pineconego.PineconeGo
	indexName       string
	projectID       *string
	namespace       string
	embedder        index.Embedder
	includeContent  bool
	batchUpsertSize int

	createIndex *CreateIndexOptions
}

type CreateIndexOptions struct {
	Dimension int
	Replicas  int
	Metric    string
	PodType   string
}

type Options struct {
	IndexName       string
	Namespace       string
	IncludeContent  bool
	BatchUpsertSize *int
	CreateIndex     *CreateIndexOptions
}

func New(options Options, embedder index.Embedder) *Index {

	apiKey := os.Getenv("PINECONE_API_KEY")
	environment := os.Getenv("PINECONE_ENVIRONMENT")

	pineconeClient := pineconego.New(environment, apiKey)

	batchUpsertSize := defaultBatchUpsertSize
	if options.BatchUpsertSize != nil {
		batchUpsertSize = *options.BatchUpsertSize
	}

	return &Index{
		pineconeClient:  pineconeClient,
		indexName:       options.IndexName,
		embedder:        embedder,
		namespace:       options.Namespace,
		includeContent:  options.IncludeContent,
		batchUpsertSize: batchUpsertSize,
		createIndex:     options.CreateIndex,
	}
}

func (p *Index) WithAPIKeyAndEnvironment(apiKey, environment string) *Index {
	p.pineconeClient = pineconego.New(environment, apiKey)
	return p
}

func (p *Index) LoadFromDocuments(ctx context.Context, documents []document.Document) error {

	err := p.createIndexIfRequired(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	err = p.batchUpsert(ctx, documents)
	if err != nil {
		return fmt.Errorf("%s: %w", index.ErrInternal, err)
	}
	return nil
}

func (p *Index) IsEmpty(ctx context.Context) (bool, error) {

	err := p.createIndexIfRequired(ctx)
	if err != nil {
		return true, fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	err = p.getProjectID(ctx)
	if err != nil {
		return true, fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	req := &pineconerequest.VectorDescribeIndexStats{
		IndexName: p.indexName,
		ProjectID: *p.projectID,
	}
	res := &pineconeresponse.VectorDescribeIndexStats{}

	err = p.pineconeClient.VectorDescribeIndexStats(ctx, req, res)
	if err != nil {
		return true, fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	namespace, ok := res.Namespaces[p.namespace]
	if !ok {
		return true, nil
	}

	if namespace.VectorCount == nil {
		return false, fmt.Errorf("%s: failed to get total index size", index.ErrInternal)
	}

	return *namespace.VectorCount == 0, nil

}

func (p *Index) SimilaritySearch(ctx context.Context, query string, opts ...option.Option) (index.SearchResponses, error) {

	pineconeOptions := &option.Options{
		TopK: defaultTopK,
	}

	for _, opt := range opts {
		opt(pineconeOptions)
	}

	if pineconeOptions.Filter == nil {
		pineconeOptions.Filter = map[string]string{}
	}

	matches, err := p.similaritySearch(ctx, query, pineconeOptions)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	searchResponses := buildSearchReponsesFromPineconeMatches(matches, p.includeContent)

	return index.FilterSearchResponses(searchResponses, pineconeOptions.TopK), nil
}

func (p *Index) similaritySearch(ctx context.Context, query string, opts *option.Options) ([]pineconeresponse.QueryMatch, error) {

	err := p.getProjectID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	embeddings, err := p.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}

	includeMetadata := true
	res := &pineconeresponse.VectorQuery{}
	err = p.pineconeClient.VectorQuery(
		ctx,
		&pineconerequest.VectorQuery{
			IndexName:       p.indexName,
			ProjectID:       *p.projectID,
			TopK:            int32(opts.TopK),
			Vector:          embeddings[0],
			IncludeMetadata: &includeMetadata,
			Namespace:       &p.namespace,
			Filter:          opts.Filter.(map[string]string),
		},
		res,
	)
	if err != nil {
		return nil, err
	}

	return res.Matches, nil
}

func (p *Index) getProjectID(ctx context.Context) error {

	if p.projectID != nil {
		return nil
	}

	whoamiResp := &pineconeresponse.Whoami{}

	err := p.pineconeClient.Whoami(ctx, &pineconerequest.Whoami{}, whoamiResp)
	if err != nil {
		return err
	}

	p.projectID = &whoamiResp.ProjectID

	return nil
}

func (p *Index) createIndexIfRequired(ctx context.Context) error {

	if p.createIndex == nil {
		return nil
	}

	resp := &pineconeresponse.IndexList{}
	err := p.pineconeClient.IndexList(ctx, &pineconerequest.IndexList{}, resp)
	if err != nil {
		return err
	}

	for _, index := range resp.Indexes {
		if index == p.indexName {
			return nil
		}
	}

	metric := pineconerequest.Metric(p.createIndex.Metric)

	req := &pineconerequest.IndexCreate{
		Name:      p.indexName,
		Dimension: p.createIndex.Dimension,
		Replicas:  &p.createIndex.Replicas,
		Metric:    &metric,
		PodType:   &p.createIndex.PodType,
	}

	err = p.pineconeClient.IndexCreate(ctx, req, &pineconeresponse.IndexCreate{})
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

			describe := &pineconeresponse.IndexDescribe{}
			err = p.pineconeClient.IndexDescribe(ctx, &pineconerequest.IndexDescribe{IndexName: p.indexName}, describe)
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

func (p *Index) batchUpsert(ctx context.Context, documents []document.Document) error {

	for i := 0; i < len(documents); i += p.batchUpsertSize {

		batchEnd := i + p.batchUpsertSize
		if batchEnd > len(documents) {
			batchEnd = len(documents)
		}

		texts := []string{}
		for _, document := range documents[i:batchEnd] {
			texts = append(texts, document.Content)
		}

		embeddings, err := p.embedder.Embed(ctx, texts)
		if err != nil {
			return err
		}

		vectors, err := buildPineconeVectorsFromEmbeddingsAndDocuments(embeddings, documents, i, p.includeContent)
		if err != nil {
			return err
		}

		err = p.vectorUpsert(ctx, vectors)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Index) vectorUpsert(ctx context.Context, vectors []pineconerequest.Vector) error {

	err := p.getProjectID(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", index.ErrInternal, err)
	}

	req := &pineconerequest.VectorUpsert{
		IndexName: p.indexName,
		ProjectID: *p.projectID,
		Vectors:   vectors,
		Namespace: p.namespace,
	}
	res := &pineconeresponse.VectorUpsert{}

	err = p.pineconeClient.VectorUpsert(ctx, req, res)
	if err != nil {
		return err
	}

	if res.UpsertedCount == nil || res.UpsertedCount != nil && *res.UpsertedCount != int64(len(vectors)) {
		return fmt.Errorf("error upserting embeddings")
	}

	return nil
}

func buildPineconeVectorsFromEmbeddingsAndDocuments(
	embeddings []embedder.Embedding,
	documents []document.Document,
	startIndex int,
	includeContent bool,
) ([]pineconerequest.Vector, error) {

	var vectors []pineconerequest.Vector

	for i, embedding := range embeddings {

		metadata := index.DeepCopyMetadata(documents[startIndex+i].Metadata)

		// inject document content into vector metadata
		if includeContent {
			metadata[index.DefaultKeyContent] = documents[startIndex+i].Content
		}

		vectorID, err := uuid.NewUUID()
		if err != nil {
			return nil, err
		}

		vectors = append(vectors, pineconerequest.Vector{
			ID:       vectorID.String(),
			Values:   embedding,
			Metadata: metadata,
		})

		// inject vector ID into document metadata
		documents[startIndex+i].Metadata[index.DefaultKeyID] = vectorID.String()
	}

	return vectors, nil
}

func buildSearchReponsesFromPineconeMatches(matches []pineconeresponse.QueryMatch, includeContent bool) index.SearchResponses {
	searchResponses := make([]index.SearchResponse, len(matches))

	for i, match := range matches {

		metadata := index.DeepCopyMetadata(match.Metadata)

		content := ""
		// extract document content from vector metadata
		if includeContent {
			content = metadata[index.DefaultKeyContent].(string)
			delete(metadata, index.DefaultKeyContent)
		}

		id := ""
		if match.ID != nil {
			id = *match.ID
		}

		score := float64(0)
		if match.Score != nil {
			score = *match.Score
		}

		searchResponses[i] = index.SearchResponse{
			ID: id,
			Document: document.Document{
				Metadata: metadata,
				Content:  content,
			},
			Score: score,
		}
	}

	return searchResponses
}
