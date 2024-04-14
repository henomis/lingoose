package transformer

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/henomis/lingoose/document"
	"github.com/henomis/lingoose/types"
	"github.com/henomis/restclientgo"
)

const (
	defaultVoyageRerankEndpoint = "https://api.voyageai.com/v1"
	defaultVoyageRerankModel    = "rerank-lite-1"
	VoyageRerankScoreMetdataKey = "voyage-rerank-score"
)

type VoyageRerank struct {
	model      string
	restClient *restclientgo.RestClient
}

func NewVoyageRerank() *VoyageRerank {
	apiKey := os.Getenv("VOYAGE_API_KEY")

	return &VoyageRerank{
		restClient: restclientgo.New(defaultVoyageRerankEndpoint).WithRequestModifier(
			func(req *http.Request) *http.Request {
				req.Header.Set("Authorization", "Bearer "+apiKey)
				return req
			}),
		model: defaultVoyageRerankModel,
	}
}

func (v *VoyageRerank) WithModel(model string) *VoyageRerank {
	v.model = model
	return v
}

func (v *VoyageRerank) Rerank(
	ctx context.Context,
	query string,
	documents []document.Document,
) ([]document.Document, error) {
	resp := &voyageRerankResponse{}
	err := v.restClient.Post(
		ctx,
		&voyageRerankRequest{
			Documents: v.documentsToStringSlice(documents),
			Model:     v.model,
			Query:     query,
		},
		resp,
	)
	if err != nil {
		return nil, err
	}

	return v.rerankDocuments(documents, resp.Data), nil
}

func (v *VoyageRerank) rerankDocuments(
	documents []document.Document,
	results []voyageRerankResponseData,
) []document.Document {
	rerankedDocuments := make([]document.Document, 0)
	for _, result := range results {
		index := result.Index
		metadata := documents[index].Metadata
		if metadata == nil {
			metadata = make(types.Meta)
		}
		metadata[VoyageRerankScoreMetdataKey] = result.RelevanceScore

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

func (v *VoyageRerank) documentsToStringSlice(documents []document.Document) []string {
	strings := make([]string, len(documents))
	for i, d := range documents {
		strings[i] = d.Content
	}
	return strings
}

// API

type voyageRerankRequest struct {
	Model     string   `json:"model"`
	Documents []string `json:"documents"`
	Query     string   `json:"query"`
}

func (r *voyageRerankRequest) Path() (string, error) {
	return "/rerank", nil
}

func (r *voyageRerankRequest) Encode() (io.Reader, error) {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(jsonBytes), nil
}

func (r *voyageRerankRequest) ContentType() string {
	return "application/json"
}

type voyageRerankResponse struct {
	HTTPStatusCode    int                        `json:"-"`
	acceptContentType string                     `json:"-"`
	Object            string                     `json:"object"`
	Data              []voyageRerankResponseData `json:"data"`
	Model             string                     `json:"model"`
	RawBody           []byte                     `json:"-"`
}

type voyageRerankResponseData struct {
	Object         string  `json:"object"`
	RelevanceScore float64 `json:"relevance_score"`
	Index          int     `json:"index"`
}

func (r *voyageRerankResponse) SetAcceptContentType(contentType string) {
	r.acceptContentType = contentType
}

func (r *voyageRerankResponse) Decode(body io.Reader) error {
	return json.NewDecoder(body).Decode(r)
}

func (r *voyageRerankResponse) SetBody(body io.Reader) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	r.RawBody = b
	return nil
}

func (r *voyageRerankResponse) AcceptContentType() string {
	if r.acceptContentType != "" {
		return r.acceptContentType
	}
	return "application/json"
}

func (r *voyageRerankResponse) SetStatusCode(code int) error {
	r.HTTPStatusCode = code
	return nil
}

func (r *voyageRerankResponse) SetHeaders(_ restclientgo.Headers) error { return nil }
