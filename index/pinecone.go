package index

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/document"
	pineconego "github.com/henomis/pinecone-go"
	pineconerequest "github.com/henomis/pinecone-go/request"
	pineconeresponse "github.com/henomis/pinecone-go/response"
)

const (
	defaultPineconeTopK = 10
)

type Pinecone struct {
	pineconeClient *pineconego.PineconeGo
	indexName      string
	projectID      string
	embedder       Embedder
}

func NewPinecone(indexName, projectID string, embedder Embedder) (*Pinecone, error) {

	apiKey := os.Getenv("PINECONE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("PINECONE_API_KEY is not set")
	}

	environment := os.Getenv("PINECONE_ENVIRONMENT")
	if environment == "" {
		return nil, fmt.Errorf("PINECONE_ENVIRONMENT is not set")
	}

	pineconeClient := pineconego.New(environment, apiKey)
	return &Pinecone{
		pineconeClient: pineconeClient,
		indexName:      indexName,
		projectID:      projectID,
		embedder:       embedder,
	}, nil
}

func (s *Pinecone) LoadFromDocuments(ctx context.Context, documents []document.Document) error {

	embeddings, err := s.embedder.Embed(ctx, documents)
	if err != nil {
		return err
	}
	var vectors []pineconerequest.Vector

	for i, embedding := range embeddings {

		metadata := make(map[string]interface{})
		source, ok := documents[i].Metadata["source"]
		if ok {
			metadata["source"] = source
		}
		metadata["index"] = fmt.Sprintf("%d", embedding.Index)
		metadata["content"] = documents[i].Content

		vectorID := uuid.New()

		vectors = append(vectors, pineconerequest.Vector{
			ID:       vectorID.String(),
			Values:   embedding.Embedding,
			Metadata: metadata,
		})
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

func (s *Pinecone) Size() (int64, error) {

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

func (s *Pinecone) SimilaritySearch(ctx context.Context, query string, topK *int) ([]SearchResponse, error) {

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

		documentContent, ok := match.Metadata["content"].(string)
		if !ok {
			documentContent = ""
		}

		documentIndex := match.Metadata["index"].(string)
		if !ok {
			documentIndex = "0"
		}

		documentIndexAsInt, err := strconv.Atoi(documentIndex)
		if err != nil {
			return nil, err
		}

		documentSource := match.Metadata["source"].(string)
		if !ok {
			documentSource = ""
		}

		score := float32(0)
		if match.Score != nil {
			score = *match.Score
		}

		document := document.Document{
			Content: documentContent,
			Metadata: map[string]interface{}{
				"source": documentSource,
			},
		}

		searchResponses[i] = SearchResponse{
			Document: document,
			Score:    score,
			Index:    documentIndexAsInt,
		}
	}

	return filterSearchResponses(searchResponses, topK), nil
}
