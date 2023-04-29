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
	pineconeClient  *pineconego.PineconeGo
	indexName       string
	projectID       string
	namespace       string
	embedder        Embedder
	includeContent  bool
	batchUpsertSize int
}

type PineconeOptions struct {
	IndexName       string
	ProjectID       string
	Namespace       string
	IncludeContent  bool
	BatchUpsertSize *int
}

func NewPinecone(options PineconeOptions, embedder Embedder) (*pinecone, error) {

	apiKey := os.Getenv("PINECONE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("PINECONE_API_KEY is not set")
	}

	environment := os.Getenv("PINECONE_ENVIRONMENT")
	if environment == "" {
		return nil, fmt.Errorf("PINECONE_ENVIRONMENT is not set")
	}

	pineconeClient := pineconego.New(environment, apiKey)

	batchUpsertSize := defaultBatchUpsertSize
	if options.BatchUpsertSize != nil {
		batchUpsertSize = *options.BatchUpsertSize
	}

	return &pinecone{
		pineconeClient:  pineconeClient,
		indexName:       options.IndexName,
		projectID:       options.ProjectID,
		embedder:        embedder,
		namespace:       options.Namespace,
		includeContent:  options.IncludeContent,
		batchUpsertSize: batchUpsertSize,
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

	searchResponses := buildSearchReponsesFromMatches(matches, p.includeContent)

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
			Namespace:       &p.namespace,
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

		embeddings, err := p.embedder.Embed(ctx, documents[i:batchEnd])
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
			Values:   embedding.Embedding,
			Metadata: metadata,
		})

		// inject vector ID into document metadata
		documents[startIndex+i].Metadata[defaultKeyID] = vectorID.String()
	}

	return vectors, nil
}

func buildSearchReponsesFromMatches(matches []pineconeresponse.QueryMatch, includeContent bool) []SearchResponse {
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

		score := float32(0)
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
