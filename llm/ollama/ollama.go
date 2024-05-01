package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/henomis/lingoose/llm/cache"
	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
	"github.com/henomis/restclientgo"
)

const (
	defaultModel      = "llama2"
	ndjsonContentType = "application/x-ndjson"
	jsonContentType   = "application/json"
	defaultEndpoint   = "http://localhost:11434/api"
)

var (
	ErrOllamaChat = fmt.Errorf("ollama chat error")
)

var threadRoleToOllamaRole = map[thread.Role]string{
	thread.RoleSystem:    "system",
	thread.RoleUser:      "user",
	thread.RoleAssistant: "assistant",
}

type StreamCallbackFn func(string)

type Observer interface {
	Span(*observer.Span) (*observer.Span, error)
	SpanEnd(*observer.Span) (*observer.Span, error)
	Generation(*observer.Generation) (*observer.Generation, error)
	GenerationEnd(*observer.Generation) (*observer.Generation, error)
}

type Ollama struct {
	model            string
	temperature      float64
	restClient       *restclientgo.RestClient
	streamCallbackFn StreamCallbackFn
	cache            *cache.Cache
	name             string
	observer         Observer
	observerTraceID  string
}

func New() *Ollama {
	return &Ollama{
		restClient: restclientgo.New(defaultEndpoint),
		model:      defaultModel,
		name:       "ollama",
	}
}

func (o *Ollama) WithEndpoint(endpoint string) *Ollama {
	o.restClient.SetEndpoint(endpoint)
	return o
}

func (o *Ollama) WithModel(model string) *Ollama {
	o.model = model
	return o
}

func (o *Ollama) WithStream(callbackFn StreamCallbackFn) *Ollama {
	o.streamCallbackFn = callbackFn
	return o
}

func (o *Ollama) WithCache(cache *cache.Cache) *Ollama {
	o.cache = cache
	return o
}

func (o *Ollama) WithTemperature(temperature float64) *Ollama {
	o.temperature = temperature
	return o
}

func (c *Ollama) WithObserver(observer Observer, traceID string) *Ollama {
	c.observer = observer
	c.observerTraceID = traceID
	return c
}

func (o *Ollama) getCache(ctx context.Context, t *thread.Thread) (*cache.Result, error) {
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

func (o *Ollama) setCache(ctx context.Context, t *thread.Thread, cacheResult *cache.Result) error {
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

func (o *Ollama) Generate(ctx context.Context, t *thread.Thread) error {
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

	var span *observer.Span
	var generation *observer.Generation
	if o.observer != nil {
		span, generation, err = o.startObserveGeneration(t)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrOllamaChat, err)
		}
	}

	if o.streamCallbackFn != nil {
		err = o.stream(ctx, t, chatRequest)
	} else {
		err = o.generate(ctx, t, chatRequest)
	}
	if err != nil {
		return err
	}

	if o.observer != nil {
		err = o.stopObserveGeneration(span, generation, t)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrOllamaChat, err)
		}
	}

	if o.cache != nil {
		err = o.setCache(ctx, t, cacheResult)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrOllamaChat, err)
		}
	}

	return nil
}

func (o *Ollama) generate(ctx context.Context, t *thread.Thread, chatRequest *request) error {
	var resp response[assistantMessage]

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
		thread.NewTextContent(resp.Message.Content),
	))

	return nil
}

func (o *Ollama) stream(ctx context.Context, t *thread.Thread, chatRequest *request) error {
	var resp response[message]
	var assistantMessage string

	resp.SetAcceptContentType(ndjsonContentType)
	resp.SetStreamCallback(
		func(data []byte) error {
			var streamResponse response[message]

			err := json.Unmarshal(data, &streamResponse)
			if err != nil {
				return err
			}

			assistantMessage += streamResponse.Message.Content
			o.streamCallbackFn(streamResponse.Message.Content)

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

func (o *Ollama) startObserveGeneration(t *thread.Thread) (*observer.Span, *observer.Generation, error) {
	span, err := o.observer.Span(
		&observer.Span{
			TraceID: o.observerTraceID,
			Name:    o.name,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	generation, err := o.observer.Generation(
		&observer.Generation{
			TraceID:  o.observerTraceID,
			ParentID: span.ID,
			Name:     fmt.Sprintf("%s-%s", o.name, o.model),
			Model:    string(o.model),
			ModelParameters: types.M{
				// TODO: Add maxTokens support
				// "maxTokens":   o.maxTokens,
				"temperature": o.temperature,
			},
			Input: t.Messages,
		},
	)
	if err != nil {
		return nil, nil, err
	}
	return span, generation, nil
}

func (o *Ollama) stopObserveGeneration(
	span *observer.Span,
	generation *observer.Generation,
	t *thread.Thread,
) error {
	_, err := o.observer.SpanEnd(span)
	if err != nil {
		return err
	}

	generation.Output = t.LastMessage()
	_, err = o.observer.GenerationEnd(generation)
	return err
}
