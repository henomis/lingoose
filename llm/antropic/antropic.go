package antropic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/henomis/lingoose/llm/cache"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/restclientgo"
)

const (
	defaultModel           = "llama2"
	eventStreamContentType = "text/event-stream"
	jsonContentType        = "application/json"
	defaultEndpoint        = "https://api.anthropic.com/v1"
)

var (
	ErrOllamaChat = fmt.Errorf("ollama chat error")
)

var threadRoleToOllamaRole = map[thread.Role]string{
	thread.RoleSystem:    "system",
	thread.RoleUser:      "user",
	thread.RoleAssistant: "assistant",
}

const (
	defaultAPIVersion = "2023-06-01"
	defaultMaxTokens  = 1024
	EOS               = "\x00"
)

type StreamCallbackFn func(string)

type Antropic struct {
	model            string
	temperature      float64
	restClient       *restclientgo.RestClient
	streamCallbackFn StreamCallbackFn
	cache            *cache.Cache
	apiVersion       string
	apiKey           string
	maxTokens        int
}

func New() *Antropic {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")

	return &Antropic{
		restClient: restclientgo.New(defaultEndpoint).WithRequestModifier(
			func(req *http.Request) *http.Request {
				req.Header.Set("x-api-key", apiKey)
				req.Header.Set("anthropic-version", defaultAPIVersion)
				return req
			},
		),
		model:      defaultModel,
		apiVersion: defaultAPIVersion,
		apiKey:     apiKey,
		maxTokens:  defaultMaxTokens,
	}
}

func (o *Antropic) WithModel(model string) *Antropic {
	o.model = model
	return o
}

func (o *Antropic) WithStream(callbackFn StreamCallbackFn) *Antropic {
	o.streamCallbackFn = callbackFn
	return o
}

func (o *Antropic) WithCache(cache *cache.Cache) *Antropic {
	o.cache = cache
	return o
}

func (o *Antropic) WithTemperature(temperature float64) *Antropic {
	o.temperature = temperature
	return o
}

func (o *Antropic) WithMaxTokens(maxTokens int) *Antropic {
	o.maxTokens = maxTokens
	return o
}

func (o *Antropic) getCache(ctx context.Context, t *thread.Thread) (*cache.Result, error) {
	messages := t.UserQuery()
	cacheQuery := strings.Join(messages, "\n")
	cacheResult, err := o.cache.Get(ctx, cacheQuery)
	if err != nil {
		return cacheResult, err
	}

	t.AddMessage(thread.NewAssistantMessage().AddContent(
		thread.NewTextContent(strings.Join(cacheResult.Answer, "\n")),
	))

	return cacheResult, nil
}

func (o *Antropic) setCache(ctx context.Context, t *thread.Thread, cacheResult *cache.Result) error {
	lastMessage := t.LastMessage()

	if lastMessage.Role != thread.RoleAssistant || len(lastMessage.Contents) == 0 {
		return nil
	}

	contents := make([]string, 0)
	for _, content := range lastMessage.Contents {
		if content.Type == thread.ContentTypeText {
			contents = append(contents, content.Data.(string))
		} else {
			contents = make([]string, 0)
			break
		}
	}

	err := o.cache.Set(ctx, cacheResult.Embedding, strings.Join(contents, "\n"))
	if err != nil {
		return err
	}

	return nil
}

func (o *Antropic) Generate(ctx context.Context, t *thread.Thread) error {
	if t == nil {
		return nil
	}

	var err error
	var cacheResult *cache.Result
	if o.cache != nil {
		cacheResult, err = o.getCache(ctx, t)
		if err == nil {
			return nil
		} else if !errors.Is(err, cache.ErrCacheMiss) {
			return fmt.Errorf("%w: %w", ErrOllamaChat, err)
		}
	}

	chatRequest := o.buildChatCompletionRequest(t)

	if o.streamCallbackFn != nil {
		err = o.stream(ctx, t, chatRequest)
	} else {
		err = o.generate(ctx, t, chatRequest)
	}

	if err != nil {
		return err
	}

	if o.cache != nil {
		err = o.setCache(ctx, t, cacheResult)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrOllamaChat, err)
		}
	}

	return nil
}

func (o *Antropic) generate(ctx context.Context, t *thread.Thread, chatRequest *request) error {
	var resp response

	err := o.restClient.Post(
		ctx,
		chatRequest,
		&resp,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOllamaChat, err)
	}

	m := thread.NewAssistantMessage()

	for _, content := range resp.Content {
		if content.Type == messageTypeText && content.Text != nil {
			m.AddContent(
				thread.NewTextContent(*content.Text),
			)
		}
	}

	t.AddMessage(m)

	return nil
}

func (o *Antropic) stream(ctx context.Context, t *thread.Thread, chatRequest *request) error {
	var resp response
	var assistantMessage string

	resp.SetAcceptContentType(eventStreamContentType)
	resp.SetStreamCallback(
		func(data []byte) error {
			dataAsString := string(data)
			if !strings.HasPrefix(dataAsString, "data: ") {
				return nil
			}

			dataAsString = strings.Replace(dataAsString, "data: ", "", -1)

			var e event
			_ = json.Unmarshal([]byte(dataAsString), &e)

			if e.Type == "content_block_delta" {
				if e.Delta != nil {
					assistantMessage += e.Delta.Text
					o.streamCallbackFn(e.Delta.Text)
				}
			} else if e.Type == "message_stop" {
				o.streamCallbackFn(EOS)
			}

			return nil
		},
	)

	chatRequest.Stream = true

	err := o.restClient.Post(
		ctx,
		chatRequest,
		&resp,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOllamaChat, err)
	}

	if resp.HTTPStatusCode >= http.StatusBadRequest {
		return fmt.Errorf("%w: %s", ErrOllamaChat, resp.RawBody)
	}

	t.AddMessage(thread.NewAssistantMessage().AddContent(
		thread.NewTextContent(assistantMessage),
	))

	return nil
}
