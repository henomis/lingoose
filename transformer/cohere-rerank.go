package transformer

import (
	"context"
	"os"

	coherego "github.com/henomis/cohere-go"
	"github.com/henomis/cohere-go/model"
	"github.com/henomis/cohere-go/request"
	"github.com/henomis/cohere-go/response"
	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
)

type CohereRerankModel = model.RerankModel

const (
	defaultCohereRerankMaxChunksPerDoc = 10
	defaultCohereRerankTopN            = -1
	CohereRerankScoreMetdataKey        = "cohere-rerank-score"

	CohereRerankModelEnglishV20      CohereRerankModel = model.RerankModelEnglishV20
	CohereRerankModelMultilingualV20 CohereRerankModel = model.RerankModelMultilingualV20
	defaultCohereRerankModel                           = CohereRerankModelEnglishV20
)

type CohereRerank struct {
	client          *coherego.Client
	maxChunksPerDoc int
	topN            int
	model           CohereRerankModel
}

func NewCohereRerank() *CohereRerank {
	return &CohereRerank{
		client:          coherego.New(os.Getenv("COHERE_API_KEY")),
		maxChunksPerDoc: defaultCohereRerankMaxChunksPerDoc,
		topN:            defaultCohereRerankTopN,
		model:           defaultCohereRerankModel,
	}
}

func (c *CohereRerank) WithMaxChunksPerDoc(maxChunksPerDoc int) *CohereRerank {
	c.maxChunksPerDoc = maxChunksPerDoc
	return c
}

func (c *CohereRerank) WithAPIKey(apiKey string) *CohereRerank {
	c.client = coherego.New(apiKey)
	return c
}

func (c *CohereRerank) WithTopN(topN int) *CohereRerank {
	c.topN = topN
	return c
}

func (c *CohereRerank) WithModel(model CohereRerankModel) *CohereRerank {
	c.model = model
	return c
}

func (c *CohereRerank) Rerank(
	ctx context.Context,
	query string, documents []document.Document,
) ([]document.Document, error) {
	if c.topN == defaultCohereRerankTopN {
		c.topN = len(documents)
	}

	resp := &response.Rerank{}
	err := c.client.Rerank(
		ctx,
		&request.Rerank{
			ReturnDocuments: false,
			MaxChunksPerDoc: &c.maxChunksPerDoc,
			Query:           query,
			Documents:       c.documentsToStringSlice(documents),
			TopN:            &c.topN,
		},
		resp,
	)
	if err != nil {
		return nil, err
	}

	return c.rerankDocuments(documents, resp.Results), nil
}

func (c *CohereRerank) documentsToStringSlice(documents []document.Document) []string {
	strings := make([]string, len(documents))
	for i, d := range documents {
		strings[i] = d.Content
	}
	return strings
}

func (c *CohereRerank) rerankDocuments(
	documents []document.Document,
	results []model.RerankResult,
) []document.Document {

	rerankedDocuments := make([]document.Document, 0)
	for _, result := range results {
		index := result.Index
		metadata := documents[index].Metadata
		if metadata == nil {
			metadata = make(types.Meta)
		}
		metadata[CohereRerankScoreMetdataKey] = result.RelevanceScore

		rerankedDocuments = append(
			rerankedDocuments,
			document.Document{
				Content:  documents[index].Content,
				Metadata: metadata,
			},
		)
	}

	return rerankedDocuments
}
