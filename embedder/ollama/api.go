package ollamaembedder

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/henomis/restclientgo"
)

type request struct {
	Model   string  `json:"model"`
	Prompt  string  `json:"prompt"`
	Options options `json:"options"`
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
	HTTPStatusCode    int       `json:"-"`
	acceptContentType string    `json:"-"`
	Embedding         []float64 `json:"embedding"`
	CreatedAt         string    `json:"created_at"`
}

func (r *response) SetAcceptContentType(contentType string) {
	r.acceptContentType = contentType
}

func (r *response) Decode(body io.Reader) error {
	return json.NewDecoder(body).Decode(r)
}

func (r *response) SetBody(_ io.Reader) error {
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

type options struct {
	Temperature float64 `json:"temperature"`
}
