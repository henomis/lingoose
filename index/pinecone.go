package index

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/henomis/lingoose/document"
	pineconego "github.com/henomis/pinecone-go"
	"github.com/henomis/pinecone-go/request"
	"github.com/henomis/pinecone-go/response"
)

const (
	defaultTopK = 10
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
	var vectors []request.Vector

	for i, embedding := range embeddings {

		metadata := make(map[string]interface{})
		source, ok := documents[i].Metadata["source"]
		if ok {
			metadata["source"] = source
		}
		metadata["index"] = fmt.Sprintf("%d", embedding.Index)
		metadata["content"] = documents[i].Content

		vectors = append(vectors, request.Vector{
			ID:       fmt.Sprintf("id-%d", embedding.Index),
			Values:   embedding.Embedding,
			Metadata: metadata,
		})
	}

	req := &request.VectorUpsert{
		IndexName: s.indexName,
		ProjectID: s.projectID,
		Vectors:   vectors,
	}
	res := &response.VectorUpsert{}

	err = s.pineconeClient.VectorUpsert(ctx, req, res)
	if err != nil {
		return err
	}

	return nil
}

func (s *Pinecone) Size() (int64, error) {

	req := &request.VectorDescribeIndexStats{
		IndexName: s.indexName,
		ProjectID: s.projectID,
	}
	res := &response.VectorDescribeIndexStats{}

	err := s.pineconeClient.VectorDescribeIndexStats(context.Background(), req, res)
	if err != nil {
		return 0, err
	}

	if res.TotalVectorCount == nil {
		return 0, fmt.Errorf("TotalVectorCount is nil")
	}

	return *res.TotalVectorCount, nil
}

func (s *Pinecone) SimilaritySearch(ctx context.Context, query string, topK *int) ([]SearchResponse, error) {

	pineconeTopK := defaultTopK
	if topK != nil {
		pineconeTopK = *topK
	}

	embeddings, err := s.embedder.Embed(ctx, []document.Document{{Content: query}})
	if err != nil {
		return nil, err
	}

	includeMetadata := true
	res := &response.VectorQuery{}
	err = s.pineconeClient.VectorQuery(
		ctx,
		&request.VectorQuery{
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

		document := document.Document{
			Content: documentContent,
			Metadata: map[string]interface{}{
				"source": documentSource,
			},
		}

		searchResponses[i] = SearchResponse{
			Document: document,
			Score:    *match.Score,
			Index:    documentIndexAsInt,
		}
	}

	//sort by similarity score
	sort.Slice(searchResponses, func(i, j int) bool {
		return searchResponses[i].Score > searchResponses[j].Score
	})

	//return topK
	if topK == nil {
		return searchResponses, nil
	}

	return searchResponses[:*topK], nil
}
