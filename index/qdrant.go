package index

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/embedder"
	qdrant "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultQdrantTopK            = 10
	defaultQdrantBatchUpsertSize = 32
)

type QdrantDistance int32

const (
	QdrantDistanceUnknow QdrantDistance = 0
	QdrantDistanceCosine QdrantDistance = 1
	QdrantDistanceEuclid QdrantDistance = 2
	QdrantDistanceDot    QdrantDistance = 3
)

type Qdrant struct {
	embedder          Embedder
	endpoint          string
	collectionName    string
	batchUpsertSize   int
	includeContent    bool
	grpcClient        *grpc.ClientConn
	collectionsClient qdrant.CollectionsClient
	pointsClient      qdrant.PointsClient

	createCollection *QdrantCreateCollectionOptions
}

type QdrantOptions struct {
	Endpoint         string
	CollectionName   string
	IncludeContent   bool
	BatchUpsertSize  *int
	CreateCollection *QdrantCreateCollectionOptions
}

type QdrantCreateCollectionOptions struct {
	Size     uint64
	Distance QdrantDistance
}

func NewQdrant(options QdrantOptions, embedder Embedder) *Qdrant {

	batchUpsertSize := defaultQdrantBatchUpsertSize
	if options.BatchUpsertSize != nil {
		batchUpsertSize = *options.BatchUpsertSize
	}

	return &Qdrant{
		endpoint:         options.Endpoint,
		embedder:         embedder,
		collectionName:   options.CollectionName,
		createCollection: options.CreateCollection,
		includeContent:   options.IncludeContent,
		batchUpsertSize:  batchUpsertSize,
	}
}

func (q *Qdrant) WithGRPCClient(client *grpc.ClientConn) *Qdrant {
	q.grpcClient = client
	return q
}

func (q *Qdrant) connect(ctx context.Context) error {

	if q.grpcClient == nil {
		q.createDefaultGRPCClient(ctx)
	}

	// _, err := qdrant.NewQdrantClient(q.grpcClient).HealthCheck(ctx, &qdrant.HealthCheckRequest{})
	// if err != nil {
	// 	return fmt.Errorf("%s: %w", ErrInternal, err)
	// }

	q.collectionsClient = qdrant.NewCollectionsClient(q.grpcClient)
	q.pointsClient = qdrant.NewPointsClient(q.grpcClient)

	return nil
}

func (q *Qdrant) createDefaultGRPCClient(ctx context.Context) error {
	var conn *grpc.ClientConn
	var err error

	conn, err = grpc.DialContext(ctx, q.endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	q.grpcClient = conn
	return nil
}

func (q *Qdrant) createCollectionIfRequired(ctx context.Context) error {

	if q.createCollection == nil {
		return nil
	}

	_, err := q.collectionsClient.Create(ctx, &qdrant.CreateCollection{
		CollectionName: q.collectionName,
		VectorsConfig: &qdrant.VectorsConfig{Config: &qdrant.VectorsConfig_Params{
			Params: &qdrant.VectorParams{
				Size:     q.createCollection.Size,
				Distance: qdrant.Distance(q.createCollection.Distance),
			},
		}},
	})
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInternal, err)
	}

	return nil
}

func (q *Qdrant) LoadFromDocuments(ctx context.Context, documents []document.Document) error {

	err := q.connect(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInternal, err)
	}

	err = q.createCollectionIfRequired(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInternal, err)
	}

	err = q.batchUpsert(ctx, documents)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInternal, err)
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

		vectors, err := buildQdrantVectorsFromEmbeddingsAndDocuments(embeddings, documents, i, q.includeContent)
		if err != nil {
			return err
		}

		err = q.upsertPoints(ctx, vectors)
		if err != nil {
			return err
		}
	}

	return nil
}

func (q *Qdrant) upsertPoints(ctx context.Context, points []*qdrant.PointStruct) error {

	waitUpsert := true
	_, err := q.pointsClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: q.collectionName,
		Wait:           &waitUpsert,
		Points:         points,
	})

	return err
}

func buildQdrantVectorsFromEmbeddingsAndDocuments(
	embeddings []embedder.Embedding,
	documents []document.Document,
	startIndex int,
	includeContent bool,
) ([]*qdrant.PointStruct, error) {

	var points []*qdrant.PointStruct

	for i, embedding := range embeddings {

		metadata := deepCopyMetadata(documents[startIndex+i].Metadata)

		// inject document content into vector metadata
		if includeContent {
			metadata[defaultKeyContent] = documents[startIndex+i].Content
		}

		pointID, err := uuid.NewUUID()
		if err != nil {
			return nil, err
		}

		payload := make(map[string]*qdrant.Value)
		for k, v := range metadata {

			payload[k] = &qdrant.Value{
				Kind: &qdrant.Value_StringValue{StringValue: v.(string)},
			}
		}

		points = append(points, &qdrant.PointStruct{
			Id: &qdrant.PointId{
				PointIdOptions: &qdrant.PointId_Uuid{Uuid: pointID.String()},
			},
			Vectors: &qdrant.Vectors{
				VectorsOptions: &qdrant.Vectors_Vector{
					Vector: &qdrant.Vector{
						Data: embedding.ToFloat32(),
					},
				},
			},
			Payload: payload,
		})

		// inject vector ID into document metadata
		documents[startIndex+i].Metadata[defaultKeyID] = pointID.String()
	}

	return points, nil
}
