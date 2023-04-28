package index

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/document"
	pineconego "github.com/henomis/pinecone-go"
	pineconerequest "github.com/henomis/pinecone-go/request"
	pineconeresponse "github.com/henomis/pinecone-go/response"
)

const (
	defaultPineconeTopK = 10
)

type pinecone struct {
	pineconeClient *pineconego.PineconeGo
	indexName      string
	projectID      string
	embedder       Embedder
}

func NewPinecone(indexName, projectID string, embedder Embedder) (*pinecone, error) {

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
	}, nil
}

func (s *pinecone) LoadFromDocuments(ctx context.Context, documents []document.Document) error {

	embeddings, err := s.embedder.Embed(ctx, documents)
	if err != nil {
		return err
	}
	var vectors []pineconerequest.Vector

	for i, embedding := range embeddings {

		metadata := make(map[string]interface{})
		for key, value := range documents[i].Metadata {
			metadata[key] = value
		}

		vectorID := uuid.New().String()

		vectors = append(vectors, pineconerequest.Vector{
			ID:       vectorID,
			Values:   embedding.Embedding,
			Metadata: metadata,
		})

		documents[i].Metadata[defaultKeyID] = vectorID
	}

	req := &pineconerequest.VectorUpsert{
		IndexName: s.indexName,
		ProjectID: s.projectID,
		Vectors:   vectors,
	}
	res := &pineconeresponse.VectorUpsert{}

	err = s.pineconeClient.VectorUpsert(ctx, req, res)
	if err != nil {
		return err
	}

	if res.UpsertedCount == nil || res.UpsertedCount != nil && *res.UpsertedCount != int64(len(vectors)) {
		return fmt.Errorf("error upserting embeddings")
	}

	return nil
}

func (s *pinecone) Size() (int64, error) {

	req := &pineconerequest.VectorDescribeIndexStats{
		IndexName: s.indexName,
		ProjectID: s.projectID,
	}
	res := &pineconeresponse.VectorDescribeIndexStats{}

	err := s.pineconeClient.VectorDescribeIndexStats(context.Background(), req, res)
	if err != nil {
		return 0, err
	}

	if res.TotalVectorCount == nil {
		return 0, fmt.Errorf("failed to get total index size")
	}

	return *res.TotalVectorCount, nil
}

func (s *pinecone) SimilaritySearch(ctx context.Context, query string, topK *int) ([]SearchResponse, error) {

	pineconeTopK := defaultPineconeTopK
	if topK != nil {
		pineconeTopK = *topK
	}

	embeddings, err := s.embedder.Embed(ctx, []document.Document{{Content: query}})
	if err != nil {
		return nil, err
	}

	includeMetadata := true
	res := &pineconeresponse.VectorQuery{}
	err = s.pineconeClient.VectorQuery(
		ctx,
		&pineconerequest.VectorQuery{
			IndexName:       s.indexName,
			ProjectID:       s.projectID,
			TopK:            int32(pineconeTopK),
			Vector:          embeddings[0].Embedding,
			IncludeMetadata: &includeMetadata,
		},
		res,
	)
	if err != nil {
		return nil, err
	}

	searchResponses := make([]SearchResponse, len(res.Matches))

	for i, match := range res.Matches {

		metadata := make(map[string]interface{})
		for k, v := range match.Metadata {
			metadata[k] = v
		}

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

	return filterSearchResponses(searchResponses, topK), nil
}
