package index

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
	pineconego "github.com/henomis/pinecone-go"
	pineconerequest "github.com/henomis/pinecone-go/request"
	pineconeresponse "github.com/henomis/pinecone-go/response"
)

const (
	defaultPineconeTopK    = 10
	defaultBatchUpsertSize = 32
)

type pinecone struct {
	pineconeClient *pineconego.PineconeGo
	indexName      string
	projectID      string
	namespace      string
	embedder       Embedder
	includeContent bool
}

func NewPinecone(indexName, projectID, namespace string, embedder Embedder, includeContent bool) (*pinecone, error) {

	apiKey := os.Getenv("PINECONE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("PINECONE_API_KEY is not set")
	}

	environment := os.Getenv("PINECONE_ENVIRONMENT")
	if environment == "" {
		return nil, fmt.Errorf("PINECONE_ENVIRONMENT is not set")
	}

	pineconeClient := pineconego.New(environment, apiKey)
	return &pinecone{
		pineconeClient: pineconeClient,
		indexName:      indexName,
		projectID:      projectID,
		embedder:       embedder,
		namespace:      namespace,
		includeContent: includeContent,
	}, nil
}

func (s *pinecone) LoadFromDocuments(ctx context.Context, documents []document.Document) error {
	return s.batchUpsert(ctx, documents)
}

func (p *pinecone) IsEmpty(ctx context.Context) (bool, error) {

	req := &pineconerequest.VectorDescribeIndexStats{
		IndexName: p.indexName,
		ProjectID: p.projectID,
	}
	res := &pineconeresponse.VectorDescribeIndexStats{}

	err := p.pineconeClient.VectorDescribeIndexStats(ctx, req, res)
	if err != nil {
		return false, err
	}

	if res.TotalVectorCount == nil {
		return false, fmt.Errorf("failed to get total index size")
	}

	return *res.TotalVectorCount == 0, nil

}

func (p *pinecone) SimilaritySearch(ctx context.Context, query string, topK *int) ([]SearchResponse, error) {

	matches, err := p.similaritySearch(ctx, topK, query)
	if err != nil {
		return nil, err
	}

	searchResponses := buildSearchReponsesFromMatches(matches)

	return filterSearchResponses(searchResponses, topK), nil
}

func (p *pinecone) similaritySearch(ctx context.Context, topK *int, query string) ([]pineconeresponse.QueryMatch, error) {
	pineconeTopK := defaultPineconeTopK
	if topK != nil {
		pineconeTopK = *topK
	}

	embeddings, err := p.embedder.Embed(ctx, []document.Document{{Content: query}})
	if err != nil {
		return nil, err
	}

	includeMetadata := true
	res := &pineconeresponse.VectorQuery{}
	err = p.pineconeClient.VectorQuery(
		ctx,
		&pineconerequest.VectorQuery{
			IndexName:       p.indexName,
			ProjectID:       p.projectID,
			TopK:            int32(pineconeTopK),
			Vector:          embeddings[0].Embedding,
			IncludeMetadata: &includeMetadata,
		},
		res,
	)
	if err != nil {
		return nil, err
	}

	return res.Matches, nil
}

func (p *pinecone) batchUpsert(ctx context.Context, documents []document.Document) error {

	for i := 0; i < len(documents); i += defaultBatchUpsertSize {

		batchEnd := i + defaultBatchUpsertSize
		if batchEnd > len(documents) {
			batchEnd = len(documents)
		}

		batchDocuments := documents[i:batchEnd]

		embeddings, err := p.embedder.Embed(ctx, batchDocuments)
		if err != nil {
			return err
		}

		vectors, err := buildVectorsFromEmbeddingsAndDocuments(embeddings, batchDocuments, p.includeContent)
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

func (p *pinecone) vectorUpsert(ctx context.Context, vectors []pineconerequest.Vector) error {

	req := &pineconerequest.VectorUpsert{
		IndexName: p.indexName,
		ProjectID: p.projectID,
		Vectors:   vectors,
		Namespace: p.namespace,
	}
	res := &pineconeresponse.VectorUpsert{}

	err := p.pineconeClient.VectorUpsert(ctx, req, res)
	if err != nil {
		return err
	}

	if res.UpsertedCount == nil || res.UpsertedCount != nil && *res.UpsertedCount != int64(len(vectors)) {
		return fmt.Errorf("error upserting embeddings")
	}

	return nil
}

func deepCopyMetadata(metadata map[string]interface{}) map[string]interface{} {
	metadataCopy := make(map[string]interface{})
	for k, v := range metadata {
		metadataCopy[k] = v
	}
	return metadataCopy
}

func buildVectorsFromEmbeddingsAndDocuments(
	embeddings []embedder.Embedding,
	documents []document.Document,
	includeContent bool,
) ([]pineconerequest.Vector, error) {

	var vectors []pineconerequest.Vector

	for i, embedding := range embeddings {

		metadata := deepCopyMetadata(documents[i].Metadata)

		if includeContent {
			metadata[defaultKeyContent] = documents[i].Content
		}

		vectorID, err := uuid.NewUUID()
		if err != nil {
			return nil, err
		}

		vectors = append(vectors, pineconerequest.Vector{
			ID:       vectorID.String(),
			Values:   embedding.Embedding,
			Metadata: metadata,
		})

		documents[i].Metadata[defaultKeyID] = vectorID
	}

	return vectors, nil
}

func buildSearchReponsesFromMatches(matches []pineconeresponse.QueryMatch) []SearchResponse {
	searchResponses := make([]SearchResponse, len(matches))

	for i, match := range matches {

		metadata := deepCopyMetadata(match.Metadata)

		id := ""
		if match.ID != nil {
			id = *match.ID
		}

		score := float32(0)
		if match.Score != nil {
			score = *match.Score
		}

		searchResponses[i] = SearchResponse{
			ID: id,
			Document: document.Document{
				Metadata: metadata,
			},
			Score: score,
		}
	}

	return searchResponses
}
