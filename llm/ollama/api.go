package ollama

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

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
	return jsonContentType
}

type response[T any] struct {
	HTTPStatusCode    int    `json:"-"`
	acceptContentType string `json:"-"`
	Model             string `json:"model"`
	CreatedAt         string `json:"created_at"`
	Message           T      `json:"message"`
	Done              bool   `json:"done"`
	streamCallbackFn  restclientgo.StreamCallback
	RawBody           []byte `json:"-"`
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

func (r *response[T]) SetBody(body io.Reader) error {
	r.RawBody, _ = io.ReadAll(body)
	return nil
}

func (r *response[T]) AcceptContentType() string {
	if r.acceptContentType != "" {
		return r.acceptContentType
	}
	return jsonContentType
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

//-------------------
// VISION
//-------------------

type visionRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
	Images []string `json:"images"`
}

func (r *visionRequest) Path() (string, error) {
	return "/generate", nil
}

func (r *visionRequest) Encode() (io.Reader, error) {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(jsonBytes), nil
}

func (r *visionRequest) ContentType() string {
	return ""
}

type visionResponse struct {
	Model          string `json:"model"`
	CreatedAt      string `json:"created_at"`
	Response       string `json:"response"`
	Done           bool   `json:"done"`
	HTTPStatusCode int    `json:"-"`
	RawBody        []byte `json:"-"`
}

func (r *visionResponse) Decode(body io.Reader) error {
	return json.NewDecoder(body).Decode(r)
}

func (r *visionResponse) SetBody(body io.Reader) error {
	r.RawBody, _ = io.ReadAll(body)
	return nil
}

func (r *visionResponse) AcceptContentType() string {
	return ndjsonContentType
}

func (r *visionResponse) SetStatusCode(code int) error {
	r.HTTPStatusCode = code
	return nil
}

func (r *visionResponse) SetHeaders(_ restclientgo.Headers) error { return nil }

func (r *visionResponse) StreamCallback() restclientgo.StreamCallback {
	return func(data []byte) error {
		var resp visionResponse
		err := json.Unmarshal(data, &resp)
		if err != nil {
			return err
		}
		r.Response += resp.Response
		return nil
	}
}

func getImageDataAsBase64(imageURL string) (string, error) {
	var imageData []byte
	var err error

	if strings.HasPrefix(imageURL, "http://") || strings.HasPrefix(imageURL, "https://") {
		resp, fetchErr := http.Get(imageURL)
		if fetchErr != nil {
			return "", err
		}
		defer resp.Body.Close()

		imageData, err = io.ReadAll(resp.Body)
	} else {
		imageData, err = os.ReadFile(imageURL)
	}
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(imageData), nil
}
