package index

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
	"github.com/henomis/lingoose/types"
	pineconego "github.com/henomis/pinecone-go"
	pineconerequest "github.com/henomis/pinecone-go/request"
	pineconeresponse "github.com/henomis/pinecone-go/response"
)

const (
	defaultPineconeTopK    = 10
	defaultBatchUpsertSize = 32
)

type Pinecone struct {
	pineconeClient  *pineconego.PineconeGo
	indexName       string
	projectID       *string
	namespace       string
	embedder        Embedder
	includeContent  bool
	batchUpsertSize int

	createIndex *PineconeCreateIndexOptions
}

type PineconeCreateIndexOptions struct {
	Dimension int
	Replicas  int
	Metric    string
	PodType   string
}

type PineconeOptions struct {
	IndexName       string
	Namespace       string
	IncludeContent  bool
	BatchUpsertSize *int
	CreateIndex     *PineconeCreateIndexOptions
}

func NewPinecone(options PineconeOptions, embedder Embedder) *Pinecone {

	apiKey := os.Getenv("PINECONE_API_KEY")
	environment := os.Getenv("PINECONE_ENVIRONMENT")

	pineconeClient := pineconego.New(environment, apiKey)

	batchUpsertSize := defaultBatchUpsertSize
	if options.BatchUpsertSize != nil {
		batchUpsertSize = *options.BatchUpsertSize
	}

	return &Pinecone{
		pineconeClient:  pineconeClient,
		indexName:       options.IndexName,
		embedder:        embedder,
		namespace:       options.Namespace,
		includeContent:  options.IncludeContent,
		batchUpsertSize: batchUpsertSize,
		createIndex:     options.CreateIndex,
	}
}

func (p *Pinecone) WithAPIKeyAndEnvironment(apiKey, environment string) *Pinecone {
	p.pineconeClient = pineconego.New(environment, apiKey)
	return p
}

func (p *Pinecone) LoadFromDocuments(ctx context.Context, documents []document.Document) error {

	err := p.createIndexIfRequired(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInternal, err)
	}

	err = p.batchUpsert(ctx, documents)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInternal, err)
	}
	return nil
}

func (p *Pinecone) IsEmpty(ctx context.Context) (bool, error) {

	err := p.createIndexIfRequired(ctx)
	if err != nil {
		return true, fmt.Errorf("%s: %w", ErrInternal, err)
	}

	err = p.getProjectID(ctx)
	if err != nil {
		return true, fmt.Errorf("%s: %w", ErrInternal, err)
	}

	req := &pineconerequest.VectorDescribeIndexStats{
		IndexName: p.indexName,
		ProjectID: *p.projectID,
	}
	res := &pineconeresponse.VectorDescribeIndexStats{}

	err = p.pineconeClient.VectorDescribeIndexStats(ctx, req, res)
	if err != nil {
		return true, fmt.Errorf("%s: %w", ErrInternal, err)
	}

	namespace, ok := res.Namespaces[p.namespace]
	if !ok {
		return true, nil
	}

	if namespace.VectorCount == nil {
		return false, fmt.Errorf("%s: failed to get total index size", ErrInternal)
	}

	return *namespace.VectorCount == 0, nil

}

func (p *Pinecone) SimilaritySearch(ctx context.Context, query string, topK *int) (SearchResponses, error) {

	matches, err := p.similaritySearch(ctx, topK, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrInternal, err)
	}

	searchResponses := buildSearchReponsesFromMatches(matches, p.includeContent)

	return filterSearchResponses(searchResponses, topK), nil
}

func (p *Pinecone) similaritySearch(ctx context.Context, topK *int, query string) ([]pineconeresponse.QueryMatch, error) {

	err := p.getProjectID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrInternal, err)
	}

	pineconeTopK := defaultPineconeTopK
	if topK != nil {
		pineconeTopK = *topK
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
			TopK:            int32(pineconeTopK),
			Vector:          embeddings[0],
			IncludeMetadata: &includeMetadata,
			Namespace:       &p.namespace,
		},
		res,
	)
	if err != nil {
		return nil, err
	}

	return res.Matches, nil
}

func (p *Pinecone) getProjectID(ctx context.Context) error {

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

func (p *Pinecone) createIndexIfRequired(ctx context.Context) error {

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

func (p *Pinecone) batchUpsert(ctx context.Context, documents []document.Document) error {

	for i := 0; i < len(documents); i += defaultBatchUpsertSize {

		batchEnd := i + defaultBatchUpsertSize
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

		vectors, err := buildVectorsFromEmbeddingsAndDocuments(embeddings, documents, i, p.includeContent)
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

func (p *Pinecone) vectorUpsert(ctx context.Context, vectors []pineconerequest.Vector) error {

	err := p.getProjectID(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInternal, err)
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

func deepCopyMetadata(metadata types.Meta) types.Meta {
	metadataCopy := make(types.Meta)
	for k, v := range metadata {
		metadataCopy[k] = v
	}
	return metadataCopy
}

func buildVectorsFromEmbeddingsAndDocuments(
	embeddings []embedder.Embedding,
	documents []document.Document,
	startIndex int,
	includeContent bool,
) ([]pineconerequest.Vector, error) {

	var vectors []pineconerequest.Vector

	for i, embedding := range embeddings {

		metadata := deepCopyMetadata(documents[startIndex+i].Metadata)

		// inject document content into vector metadata
		if includeContent {
			metadata[defaultKeyContent] = documents[startIndex+i].Content
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
		documents[startIndex+i].Metadata[defaultKeyID] = vectorID.String()
	}

	return vectors, nil
}

func buildSearchReponsesFromMatches(matches []pineconeresponse.QueryMatch, includeContent bool) SearchResponses {
	searchResponses := make([]SearchResponse, len(matches))

	for i, match := range matches {

		metadata := deepCopyMetadata(match.Metadata)

		content := ""
		// extract document content from vector metadata
		if includeContent {
			content = metadata[defaultKeyContent].(string)
			delete(metadata, defaultKeyContent)
		}

		id := ""
		if match.ID != nil {
			id = *match.ID
		}

		score := float64(0)
		if match.Score != nil {
			score = *match.Score
		}

		searchResponses[i] = SearchResponse{
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
