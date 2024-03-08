package antropic

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
	Model         string    `json:"model"`
	Messages      []message `json:"messages"`
	System        string    `json:"system"`
	MaxTokens     int       `json:"max_tokens"`
	Metadata      metadata  `json:"metadata"`
	StopSequences []string  `json:"stop_sequences"`
	Stream        bool      `json:"stream"`
	Temperature   float64   `json:"temperature"`
	TopP          float64   `json:"top_p"`
	TopK          int       `json:"top_k"`
}

type metadata struct {
	UserID string `json:"user_id"`
}

func (r *request) Path() (string, error) {
	return "/messages", nil
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

type response struct {
	HTTPStatusCode    int       `json:"-"`
	acceptContentType string    `json:"-"`
	ID                string    `json:"id"`
	Type              string    `json:"type"`
	Error             aerror    `json:"error"`
	Role              string    `json:"role"`
	Content           []content `json:"content"`
	Model             string    `json:"model"`
	StopReason        *string   `json:"stop_reason"`
	StopSequence      *string   `json:"stop_sequence"`
	Usage             usage     `json:"usage"`
	streamCallbackFn  restclientgo.StreamCallback
	RawBody           []byte `json:"-"`
}

type aerror struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type content struct {
	Type   contentType    `json:"type"`
	Text   *string        `json:"text,omitempty"`
	Source *contentSource `json:"source,omitempty"`
}

type contentSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (r *response) SetAcceptContentType(contentType string) {
	r.acceptContentType = contentType
}

func (r *response) Decode(body io.Reader) error {
	return json.NewDecoder(body).Decode(r)
}

func (r *response) SetBody(body io.Reader) error {
	r.RawBody, _ = io.ReadAll(body)
	return nil
}

func (r *response) AcceptContentType() string {
	if r.acceptContentType != "" {
		return r.acceptContentType
	}
	return jsonContentType
}

func (r *response) SetStatusCode(code int) error {
	r.HTTPStatusCode = code
	return nil
}

func (r *response) SetHeaders(_ restclientgo.Headers) error { return nil }

func (r *response) SetStreamCallback(fn restclientgo.StreamCallback) {
	r.streamCallbackFn = fn
}

func (r *response) StreamCallback() restclientgo.StreamCallback {
	return r.streamCallbackFn
}

type message struct {
	Role    string    `json:"role"`
	Content []content `json:"content"`
}

type contentType string

const (
	messageTypeText  contentType = "text"
	messageTypeImage contentType = "image"
)

// {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "Hello"}}

type event struct {
	Type  string `json:"type"`
	Index *int   `json:"index,omitempty"`
	Delta *delta `json:"delta,omitempty"`
}

type delta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func getImageDataAsBase64(imageURL string) (string, string, error) {
	var imageData []byte
	var err error

	if strings.HasPrefix(imageURL, "http://") || strings.HasPrefix(imageURL, "https://") {
		//nolint:gosec
		resp, fetchErr := http.Get(imageURL)
		if fetchErr != nil {
			return "", "", fetchErr
		}
		defer resp.Body.Close()

		imageData, err = io.ReadAll(resp.Body)
	} else {
		imageData, err = os.ReadFile(imageURL)
	}
	if err != nil {
		return "", "", err
	}

	// Detect image type
	mimeType := http.DetectContentType(imageData)

	return base64.StdEncoding.EncodeToString(imageData), mimeType, nil
}
