package ollama

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/henomis/restclientgo"
)

type request struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
	Options  options   `json:"options"`
}

func (r *request) Path() (string, error) {
	return "/chat", nil
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

type response[T any] struct {
	HTTPStatusCode    int    `json:"-"`
	acceptContentType string `json:"-"`
	Model             string `json:"model"`
	CreatedAt         string `json:"created_at"`
	Message           T      `json:"message"`
	Done              bool   `json:"done"`
	streamCallbackFn  restclientgo.StreamCallback
}

type assistantMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (r *response[T]) SetAcceptContentType(contentType string) {
	r.acceptContentType = contentType
}

func (r *response[T]) Decode(body io.Reader) error {
	return json.NewDecoder(body).Decode(r)
}

func (r *response[T]) SetBody(_ io.Reader) error {
	return nil
}

func (r *response[T]) AcceptContentType() string {
	if r.acceptContentType != "" {
		return r.acceptContentType
	}
	return "application/json"
}

func (r *response[T]) SetStatusCode(code int) error {
	r.HTTPStatusCode = code
	return nil
}

func (r *response[T]) SetHeaders(_ restclientgo.Headers) error { return nil }

func (r *response[T]) SetStreamCallback(fn restclientgo.StreamCallback) {
	r.streamCallbackFn = fn
}

func (r *response[T]) StreamCallback() restclientgo.StreamCallback {
	return r.streamCallbackFn
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type options struct {
	Temperature float64 `json:"temperature"`
}
