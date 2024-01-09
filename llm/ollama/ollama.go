package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/henomis/lingoose/llm/cache"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/restclientgo"
)

const (
	defaultModel      = "llama2"
	ndjsonContentType = "application/x-ndjson"
	defaultEndpoint   = "http://localhost:11434/api"
)

var (
	ErrOllamaChat = fmt.Errorf("ollama chat error")
)

var threadRoleToOpenAIRole = map[thread.Role]string{
	thread.RoleSystem:    "system",
	thread.RoleUser:      "user",
	thread.RoleAssistant: "assistant",
}

type StreamCallbackFn func(string)

type Ollama struct {
	model            string
	restClient       *restclientgo.RestClient
	streamCallbackFn StreamCallbackFn
	cache            *cache.Cache
}

func New() *Ollama {
	return &Ollama{
		restClient: restclientgo.New(defaultEndpoint),
		model:      defaultModel,
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

func (o *Ollama) WithStream(enable bool, callbackFn StreamCallbackFn) *Ollama {
	if !enable {
		o.streamCallbackFn = nil
	} else {
		o.streamCallbackFn = callbackFn
	}

	return o
}

func (o *Ollama) WithCache(cache *cache.Cache) *Ollama {
	o.cache = cache
	return o
}

func getCacheableMessages(t *thread.Thread) []string {
	userMessages := make([]*thread.Message, 0)
	for _, message := range t.Messages {
		if message.Role == thread.RoleUser {
			userMessages = append(userMessages, message)
		} else {
			userMessages = make([]*thread.Message, 0)
		}
	}

	var messages []string
	for _, message := range userMessages {
		for _, content := range message.Contents {
			if content.Type == thread.ContentTypeText {
				messages = append(messages, content.Data.(string))
			} else {
				messages = make([]string, 0)
				break
			}
		}
	}

	return messages
}

func (o *Ollama) getCache(ctx context.Context, t *thread.Thread) (*cache.Result, error) {
	messages := getCacheableMessages(t)
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
	lastMessage := t.Messages[len(t.Messages)-1]

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

func (o *Ollama) generate(ctx context.Context, t *thread.Thread, chatRequest *request) error {
	var chatResponse chatResponse

	o.restClient.SetStreamCallback(nil)

	err := o.restClient.Post(
		ctx,
		chatRequest,
		&chatResponse,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOllamaChat, err)
	}

	t.AddMessage(thread.NewAssistantMessage().AddContent(
		thread.NewTextContent(chatResponse.AssistantMessage.Content),
	))

	return nil
}

func (o *Ollama) stream(ctx context.Context, t *thread.Thread, chatRequest *request) error {
	streamChatResponse := &chatStreamResponse{}
	var assistantMessage string

	streamChatResponse.SetAcceptContentType(ndjsonContentType)
	o.restClient.SetStreamCallback(
		func(data []byte) error {
			var chatResponse chatStreamResponse

			err := json.Unmarshal(data, &chatResponse)
			if err != nil {
				return err
			}

			assistantMessage += chatResponse.Message.Content
			o.streamCallbackFn(chatResponse.Message.Content)

			return nil
		},
	)

	chatRequest.Stream = true

	err := o.restClient.Post(
		ctx,
		chatRequest,
		streamChatResponse,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOllamaChat, err)
	}

	t.AddMessage(thread.NewAssistantMessage().AddContent(
		thread.NewTextContent(assistantMessage),
	))

	return nil
}
