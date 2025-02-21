package voyageembedder

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/henomis/restclientgo"
	"github.com/rsest/lingoose/embedder"
)

type request struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

func (r *request) Path() (string, error) {
	return "/embeddings", nil
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
	HTTPStatusCode    int    `json:"-"`
	acceptContentType string `json:"-"`
	Object            string `json:"object"`
	Data              []data `json:"data"`
	Model             string `json:"model"`
	RawBody           []byte `json:"-"`
}

type data struct {
	Object    string             `json:"object"`
	Embedding embedder.Embedding `json:"embedding"`
	Index     int                `json:"index"`
}

func (r *response) SetAcceptContentType(contentType string) {
	r.acceptContentType = contentType
}

func (r *response) Decode(body io.Reader) error {
	return json.NewDecoder(body).Decode(r)
}

func (r *response) SetBody(body io.Reader) error {
	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	r.RawBody = b
	return nil
}

func (r *response) AcceptContentType() string {
	if r.acceptContentType != "" {
		return r.acceptContentType
	}
	return "application/json"
}

func (r *response) SetStatusCode(code int) error {
	r.HTTPStatusCode = code
	return nil
}

func (r *response) SetHeaders(_ restclientgo.Headers) error { return nil }
