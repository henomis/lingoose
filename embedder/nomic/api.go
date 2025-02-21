package nomicembedder

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/henomis/restclientgo"
	"github.com/rsest/lingoose/embedder"
)

type Model string

const (
	ModelNomicEmbedTextV1 Model = "nomic-embed-text-v1"
	ModelAllMiniLML6V2    Model = "all-MiniLM-L6-v2"
)

type TaskType string

const (
	TaskTypeSearchQuery    TaskType = "search_query"
	TaskTypeSearchDocument TaskType = "search_document"
	TaskTypeClustering     TaskType = "clustering"
	TaskTypeClassification TaskType = "classification"
)

type request struct {
	Model    string   `json:"model"`
	Texts    []string `json:"texts"`
	TaskType TaskType `json:"task_type,omitempty"`
}

func (r *request) Path() (string, error) {
	return "/embedding/text", nil
}

func (r *request) Encode() (io.Reader, error) {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(jsonBytes), nil
}

func (r *request) ContentType() string {
	return "application/json"
}

type response struct {
	HTTPStatusCode int                  `json:"-"`
	Embeddings     []embedder.Embedding `json:"embeddings"`
	Usage          Usage                `json:"usage"`
	RawBody        string               `json:"-"`
}

type Usage struct {
	TotalTokens int `json:"total_tokens"`
}

func (r *response) Decode(body io.Reader) error {
	return json.NewDecoder(body).Decode(r)
}

func (r *response) SetBody(body io.Reader) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	r.RawBody = string(b)
	return nil
}

func (r *response) AcceptContentType() string {
	return "application/json"
}

func (r *response) SetStatusCode(code int) error {
	r.HTTPStatusCode = code
	return nil
}

func (r *response) SetHeaders(_ restclientgo.Headers) error { return nil }
