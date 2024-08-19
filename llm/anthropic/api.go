package anthropic

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
	Model         string      `json:"model"`
	Messages      []message   `json:"messages"`
	Tools         []tool      `json:"tools,omitempty"`
	ToolChoice    *toolChoice `json:"tool_choice,omitempty"`
	System        string      `json:"system"`
	MaxTokens     int         `json:"max_tokens"`
	Metadata      metadata    `json:"metadata"`
	StopSequences []string    `json:"stop_sequences"`
	Stream        bool        `json:"stream"`
	Temperature   float64     `json:"temperature"`
	TopP          float64     `json:"top_p"`
	TopK          int         `json:"top_k"`
}

type tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"input_schema"`
}

type toolChoice struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
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
	Type      contentType     `json:"type"`
	Text      *string         `json:"text,omitempty"`
	Source    *contentSource  `json:"source,omitempty"`
	Id        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseId string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
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

type contentType string
type eventType string

const (
	messageTypeText            contentType = "text"
	messageTypeImage           contentType = "image"
	messageTypeToolUse         contentType = "tool_use"
	messageTypeToolResult      contentType = "tool_result"
	eventTypeMessageStart      eventType   = "message_start"
	eventTypeMessageDelta      eventType   = "message_delta"
	eventTypeMessageStop       eventType   = "message_stop"
	eventTypePing              eventType   = "ping"
	eventTypeContentBlockStart eventType   = "content_block_start"
	eventTypeContentBlockDelta eventType   = "content_block_delta"
	eventTypeContentBlockStop  eventType   = "content_block_stop"
)

type message struct {
	Id           string    `json:"id,omitempty"`
	Type         string    `json:"type,omitempty"`
	Role         string    `json:"role"`
	Model        string    `json:"model,omitempty"`
	Content      []content `json:"content"`
	StopReason   *string   `json:"stop_reason,omitempty"`
	StopSequence *string   `json:"stop_sequence,omitempty"`
	Usage        *usage    `json:"usage,omitempty"`
}

type event struct {
	Type         eventType `json:"type"`
	Index        *int      `json:"index,omitempty"`
	Message      *message  `json:"message,omitempty"`
	Delta        *delta    `json:"delta,omitempty"`
	ContentBlock *content  `json:"content_block,omitempty"`
	Usage        *usage    `json:"usage,omitempty"`
}

type delta struct {
	Type         string          `json:"type"`
	Text         string          `json:"text,omitempty"`
	PartialJson  json.RawMessage `json:"partial_json,omitempty"`
	StopReason   *string         `json:"stop_reason"`
	StopSequence *string         `json:"stop_sequence"`
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