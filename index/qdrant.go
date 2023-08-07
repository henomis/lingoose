package index

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
	qdrantgo "github.com/henomis/qdrant-go"
	qdrantrequest "github.com/henomis/qdrant-go/request"
	qdrantresponse "github.com/henomis/qdrant-go/response"
)

const (
	defaultQdrantTopK            = 10
	defaultQdrantBatchUpsertSize = 32
)

type Qdrant struct {
	qdrantClient    *qdrantgo.Client
	collectionName  string
	embedder        Embedder
	includeContent  bool
	batchUpsertSize int

	createCollection *QdrantCreateCollectionOptions
}

type QdrantDistance string

const (
	QdrantDistanceCosine    QdrantDistance = QdrantDistance(qdrantrequest.DistanceCosine)
	QdrantDistanceEuclidean QdrantDistance = QdrantDistance(qdrantrequest.DistanceEuclidean)
	QdrantDistanceDot       QdrantDistance = QdrantDistance(qdrantrequest.DistanceDot)
)

type QdrantCreateCollectionOptions struct {
	Dimension uint64
	Distance  QdrantDistance
	OnDisk    bool
}

type QdrantOptions struct {
	CollectionName   string
	IncludeContent   bool
	BatchUpsertSize  *int
	CreateCollection *QdrantCreateCollectionOptions
}

func NewQdrant(options QdrantOptions, embedder Embedder) *Qdrant {

	apiKey := os.Getenv("QDRANT_API_KEY")
	endpoint := os.Getenv("QDRANT_ENDPOINT")

	qdrantClient := qdrantgo.New(endpoint, apiKey)

	batchUpsertSize := defaultQdrantBatchUpsertSize
	if options.BatchUpsertSize != nil {
		batchUpsertSize = *options.BatchUpsertSize
	}

	return &Qdrant{
		qdrantClient:     qdrantClient,
		collectionName:   options.CollectionName,
		embedder:         embedder,
		includeContent:   options.IncludeContent,
		batchUpsertSize:  batchUpsertSize,
		createCollection: options.CreateCollection,
	}
}

func (q *Qdrant) WithAPIKeyAndEdpoint(apiKey, endpoint string) *Qdrant {
	q.qdrantClient = qdrantgo.New(endpoint, apiKey)
	return q
}

func (q *Qdrant) LoadFromDocuments(ctx context.Context, documents []document.Document) error {

	err := q.createCollectionIfRequired(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInternal, err)
	}

	err = q.batchUpsert(ctx, documents)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInternal, err)
	}
	return nil
}

func (p *Qdrant) IsEmpty(ctx context.Context) (bool, error) {

	err := p.createCollectionIfRequired(ctx)
	if err != nil {
		return true, fmt.Errorf("%s: %w", ErrInternal, err)
	}

	res := &qdrantresponse.CollectionCollectInfo{}
	err = p.qdrantClient.CollectionCollectInfo(
		ctx,
		&qdrantrequest.CollectionCollectInfo{
			CollectionName: p.collectionName,
		},
		res,
	)
	if err != nil {
		return true, fmt.Errorf("%s: %w", ErrInternal, err)
	}

	return res.Result.VectorsCount == 0, nil

}

func (q *Qdrant) SimilaritySearch(ctx context.Context, query string, opts ...Option) (SearchResponses, error) {

	qdrantOptions := &options{
		topK: defaultQdrantTopK,
	}

	for _, opt := range opts {
		opt(qdrantOptions)
	}

	matches, err := q.similaritySearch(ctx, query, qdrantOptions)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrInternal, err)
	}

	searchResponses := buildSearchReponsesFromQdrantMatches(matches, q.includeContent)

	return filterSearchResponses(searchResponses, qdrantOptions.topK), nil
}

func (p *Qdrant) similaritySearch(ctx context.Context, query string, opts *options) ([]qdrantresponse.PointSearchResult, error) {

	embeddings, err := p.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}

	includeMetadata := true
	res := &qdrantresponse.PointSearch{}
	err = p.qdrantClient.PointSearch(
		ctx,
		&qdrantrequest.PointSearch{
			CollectionName: p.collectionName,
			Limit:          opts.topK,
			Vector:         embeddings[0],
			WithPayload:    &includeMetadata,
			Filter:         opts.filter.(qdrantrequest.Filter),
		},
		res,
	)
	if err != nil {
		return nil, err
	}

	return res.Result, nil
}

func (q *Qdrant) createCollectionIfRequired(ctx context.Context) error {

	if q.createCollection == nil {
		return nil
	}

	resp := &qdrantresponse.CollectionList{}
	err := q.qdrantClient.CollectionList(ctx, &qdrantrequest.CollectionList{}, resp)
	if err != nil {
		return err
	}

	for _, collection := range resp.Result.Collections {
		if collection.Name == q.collectionName {
			return nil
		}
	}

	req := &qdrantrequest.CollectionCreate{
		CollectionName: q.collectionName,
		Vectors: qdrantrequest.VectorsParams{
			Size:     q.createCollection.Dimension,
			Distance: qdrantrequest.Distance(q.createCollection.Distance),
			OnDisk:   &q.createCollection.OnDisk,
		},
	}

	err = q.qdrantClient.CollectionCreate(ctx, req, &qdrantresponse.CollectionCreate{})
	if err != nil {
		return err
	}

	return nil
}

func (q *Qdrant) batchUpsert(ctx context.Context, documents []document.Document) error {

	for i := 0; i < len(documents); i += q.batchUpsertSize {

		batchEnd := i + q.batchUpsertSize
		if batchEnd > len(documents) {
			batchEnd = len(documents)
		}

		texts := []string{}
		for _, document := range documents[i:batchEnd] {
			texts = append(texts, document.Content)
		}

		embeddings, err := q.embedder.Embed(ctx, texts)
		if err != nil {
			return err
		}

		points, err := buildQdrantPointsFromEmbeddingsAndDocuments(embeddings, documents, i, q.includeContent)
		if err != nil {
			return err
		}

		err = q.pointUpsert(ctx, points)
		if err != nil {
			return err
		}
	}

	return nil
}

func (q *Qdrant) pointUpsert(ctx context.Context, points []qdrantrequest.Point) error {

	wait := true
	req := &qdrantrequest.PointUpsert{
		Wait:           &wait,
		CollectionName: q.collectionName,
		Points:         points,
	}
	res := &qdrantresponse.PointUpsert{}

	err := q.qdrantClient.PointUpsert(ctx, req, res)
	if err != nil {
		return err
	}

	return nil
}

func buildQdrantPointsFromEmbeddingsAndDocuments(
	embeddings []embedder.Embedding,
	documents []document.Document,
	startIndex int,
	includeContent bool,
) ([]qdrantrequest.Point, error) {

	var vectors []qdrantrequest.Point

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

		vectors = append(vectors, qdrantrequest.Point{
			ID:      vectorID.String(),
			Vector:  embedding,
			Payload: metadata,
		})

		// inject vector ID into document metadata
		documents[startIndex+i].Metadata[defaultKeyID] = vectorID.String()
	}

	return vectors, nil
}

func buildSearchReponsesFromQdrantMatches(matches []qdrantresponse.PointSearchResult, includeContent bool) SearchResponses {
	searchResponses := make([]SearchResponse, len(matches))

	for i, match := range matches {

		metadata := deepCopyMetadata(match.Payload)

		content := ""
		// extract document content from vector metadata
		if includeContent {
			content = metadata[defaultKeyContent].(string)
			delete(metadata, defaultKeyContent)
		}

		searchResponses[i] = SearchResponse{
			ID: match.ID,
			Document: document.Document{
				Metadata: metadata,
				Content:  content,
			},
			Score: match.Score,
		}
	}

	return searchResponses
}
