package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"os"
	"strings"

	"github.com/henomis/lingoose/llm/cache"
	llmobserver "github.com/henomis/lingoose/llm/observer"
	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
	"github.com/henomis/restclientgo"
)

const (
	defaultModel           = "claude-3-opus-20240229"
	eventStreamContentType = "text/event-stream"
	jsonContentType        = "application/json"
	defaultEndpoint        = "https://api.anthropic.com/v1"
)

var (
	ErrAnthropicChat = fmt.Errorf("anthropic chat error")
)

var threadRoleToAnthropicRole = map[thread.Role]string{
	thread.RoleSystem:    "system",
	thread.RoleUser:      "user",
	thread.RoleAssistant: "assistant",
}

const (
	defaultAPIVersion  = "2023-06-01"
	defaultBetaVersion = "tools-2024-04-04"
	defaultMaxTokens   = 1024
	EOS                = "\x00"
)

type Model string

const (
	Claude3Opus20240229   Model = "claude-3-opus-20240229"
	Claude3Opus           Model = "claude-3-opus-20240229"
	Claude3Sonnet20240229 Model = "claude-3-sonnet-20240229"
	Claude3Sonnet         Model = "claude-3-sonnet-20240229"
	Claude3Haiku20240307  Model = "claude-3-haiku-20240307"
	Claude3Haiku          Model = "claude-3-haiku-20240307"
	Claude3Dot5Sonnet     Model = "claude-3-5-sonnet-20240620"
)

type UsageCallback func(types.Meta)
type StreamCallbackFn func(string)

type Anthropic struct {
	model            Model
	temperature      float64
	restClient       *restclientgo.RestClient
	streamCallbackFn StreamCallbackFn
	usageCallback    UsageCallback
	cache            *cache.Cache
	apiVersion       string
	apiKey           string
	maxTokens        int
	name             string
	functions        map[string]Function
}

func New() *Anthropic {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")

	return &Anthropic{
		restClient: restclientgo.New(defaultEndpoint).WithRequestModifier(
			func(req *http.Request) *http.Request {
				req.Header.Set("x-api-key", apiKey)
				req.Header.Set("anthropic-version", defaultAPIVersion)
				//req.Header.Set("anthropic-beta", defaultBetaVersion)
				return req
			},
		),
		model:      defaultModel,
		apiVersion: defaultAPIVersion,
		apiKey:     apiKey,
		maxTokens:  defaultMaxTokens,
		functions:  make(map[string]Function),
	}
}

func (o *Anthropic) WithModel(model Model) *Anthropic {
	o.model = model
	return o
}

func (o *Anthropic) WithStream(callbackFn StreamCallbackFn) *Anthropic {
	o.streamCallbackFn = callbackFn
	return o
}

func (o *Anthropic) WithCache(cache *cache.Cache) *Anthropic {
	o.cache = cache
	return o
}

func (o *Anthropic) WithUsageCallback(callback UsageCallback) *Anthropic {
	o.usageCallback = callback
	return o
}

func (o *Anthropic) WithTemperature(temperature float64) *Anthropic {
	o.temperature = temperature
	return o
}

func (o *Anthropic) WithMaxTokens(maxTokens int) *Anthropic {
	o.maxTokens = maxTokens
	return o
}

func (o *Anthropic) getCache(ctx context.Context, t *thread.Thread) (*cache.Result, error) {
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

func (o *Anthropic) setCache(ctx context.Context, t *thread.Thread, cacheResult *cache.Result) error {
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

func (o *Anthropic) Generate(ctx context.Context, t *thread.Thread) error {
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
			return fmt.Errorf("%w: %w", ErrAnthropicChat, err)
		}
	}

	chatRequest := o.buildChatCompletionRequest(t)

	generation, err := o.startObserveGeneration(ctx, t)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAnthropicChat, err)
	}

	if o.streamCallbackFn != nil {
		err = o.stream(ctx, t, chatRequest)
	} else {
		err = o.generate(ctx, t, chatRequest)
	}
	if err != nil {
		return err
	}

	err = o.stopObserveGeneration(ctx, generation, t)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAnthropicChat, err)
	}

	if o.cache != nil {
		err = o.setCache(ctx, t, cacheResult)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrAnthropicChat, err)
		}
	}

	return nil
}

func (o *Anthropic) generate(ctx context.Context, t *thread.Thread, chatRequest *request) error {
	var resp response

	err := o.restClient.Post(
		ctx,
		chatRequest,
		&resp,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAnthropicChat, err)
	}

	m := thread.NewAssistantMessage()
	mr := thread.NewUserMessage()

	for _, c := range resp.Content {
		if c.Type == messageTypeText && c.Text != nil {
			m.AddContent(
				thread.NewTextContent(*c.Text),
			)
		} else if c.Type == messageTypeToolUse {
			m.AddContent(toolCallsToToolCallContent(c))
			toolResponse, err := o.callTool(c)
			if err != nil {
				return err
			}
			mr.AddContent(toolResponse)
		}
	}

	t.AddMessage(m)
	if mr.Contents != nil {
		t.AddMessages(mr)
	}

	return nil
}

func (o *Anthropic) setUsageMetadata(usage usage) {
	callbackMetadata := make(types.Meta)

	err := mapstructure.Decode(usage, &callbackMetadata)
	if err != nil {
		return
	}

	o.usageCallback(callbackMetadata)
}

func (o *Anthropic) stream(ctx context.Context, t *thread.Thread, chatRequest *request) error {
	var resp response
	var assistantMessage string
	var currentToolCall *content

	var messages []*thread.Message
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

			switch e.Type {
			case eventTypeMessageStart:
				o.setUsageMetadata(*e.Message.Usage)
			case eventTypeContentBlockStart:
				if e.ContentBlock.Type == messageTypeToolUse {
					currentToolCall = e.ContentBlock
				}
			case eventTypeContentBlockDelta:
				if e.Delta != nil {
					if e.Delta.PartialJson != nil {
						if string(currentToolCall.Input) == "{}" {
							currentToolCall.Input = currentToolCall.Input[:0]
						}
						partialArgs := strings.TrimSuffix(strings.TrimPrefix(string(e.Delta.PartialJson), `"`), `"`)
						currentToolCall.Input = append(currentToolCall.Input, []byte(partialArgs)...)
					}
					if e.Delta.Text != "" {
						assistantMessage += e.Delta.Text
						o.streamCallbackFn(e.Delta.Text)
					}
				}
			case eventTypeContentBlockStop:
				if currentToolCall != nil {
					var unquotedArgs string
					currentToolCall.Input = []byte(`"` + string(currentToolCall.Input) + `"`)
					err := json.Unmarshal(currentToolCall.Input, &unquotedArgs)
					if err == nil {
						currentToolCall.Input = []byte(unquotedArgs)
					}
					messages = append(messages, thread.NewAssistantMessage().AddContent(toolCallsToToolCallContent(*currentToolCall)))
					toolResponse, err := o.callTool(*currentToolCall)
					if err != nil {
						return err
					}
					messages = append(messages, thread.NewUserMessage().AddContent(toolResponse))
					currentToolCall = nil
				}
				if assistantMessage != "" {
					messages = append(messages, thread.NewAssistantMessage().AddContent(thread.NewTextContent(assistantMessage)))
					assistantMessage = ""
				}
			case eventTypeMessageDelta:
				if e.Usage != nil {
					o.setUsageMetadata(*e.Usage)
				}
			case eventTypeMessageStop:
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
		return fmt.Errorf("%w: %w", ErrAnthropicChat, err)
	}

	if resp.HTTPStatusCode >= http.StatusBadRequest {
		return fmt.Errorf("%w: %s", ErrAnthropicChat, resp.RawBody)
	}

	t.AddMessages(messages...)
	return nil
}

func (o *Anthropic) callTool(toolCall content) (*thread.Content, error) {
	fn, ok := o.functions[toolCall.Name]
	if !ok {
		return nil, fmt.Errorf("unknown function %s", toolCall.Name)
	}

	toolResponseData := thread.ToolResponseData{
		ID:   toolCall.Id,
		Name: toolCall.Name,
	}
	resultAsJSON, err := callFnWithArgumentAsJSON(fn.Fn, string(toolCall.Input))
	if err != nil {
		toolResponseData.Result = "Error: " + err.Error()
	} else {
		toolResponseData.Result = resultAsJSON
	}
	return thread.NewToolResponseContent(toolResponseData), nil
}

func (o *Anthropic) startObserveGeneration(ctx context.Context, t *thread.Thread) (*observer.Generation, error) {
	return llmobserver.StartObserveGeneration(
		ctx,
		o.name,
		string(o.model),
		types.M{
			"maxTokens":   o.maxTokens,
			"temperature": o.temperature,
		},
		t,
	)
}

func (o *Anthropic) stopObserveGeneration(
	ctx context.Context,
	generation *observer.Generation,
	t *thread.Thread,
) error {
	return llmobserver.StopObserveGeneration(
		ctx,
		generation,
		t,
	)
}
