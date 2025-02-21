package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/henomis/restclientgo"

	"github.com/rsest/lingoose/llm/cache"
	llmobserver "github.com/rsest/lingoose/llm/observer"
	"github.com/rsest/lingoose/observer"
	"github.com/rsest/lingoose/thread"
	"github.com/rsest/lingoose/types"
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

type Ollama struct {
	model            string
	temperature      float64
	restClient       *restclientgo.RestClient
	streamCallbackFn StreamCallbackFn
	cache            *cache.Cache
	name             string
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

	generation, err := o.startObserveGeneration(ctx, t)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOllamaChat, err)
	}

	if o.streamCallbackFn != nil {
		err = o.stream(ctx, t, chatRequest)
	} else {
		err = o.generate(ctx, t, chatRequest)
	}
	if err != nil {
		return err
	}

	err = o.stopObserveGeneration(ctx, generation, []*thread.Message{t.LastMessage()})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOllamaChat, err)
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

func (o *Ollama) startObserveGeneration(ctx context.Context, t *thread.Thread) (*observer.Generation, error) {
	return llmobserver.StartObserveGeneration(
		ctx,
		o.name,
		o.model,
		types.M{
			// TODO: Add maxTokens parameter
			// "maxTokens":   o.maxTokens,
			"temperature": o.temperature,
		},
		t,
	)
}

func (o *Ollama) stopObserveGeneration(
	ctx context.Context,
	generation *observer.Generation,
	messages []*thread.Message,
) error {
	return llmobserver.StopObserveGeneration(
		ctx,
		generation,
		messages,
	)
}
