package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/henomis/lingoose/llm/cache"
	llmobserver "github.com/henomis/lingoose/llm/observer"
	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

const (
	// EOS represents the end of stream marker
	EOS = "\x00"
	// deltaTypeText represents a text content delta in the stream
	deltaTypeText = "text"
	// deltaTypeToolUse represents a tool use content delta in the stream
	deltaTypeToolUse = "tool_use"
)

// Anthropic represents the main client structure.
type Anthropic struct {
	client           *anthropicsdk.Client
	model            Model
	temperature      float64
	maxTokens        int
	stop             []string
	usageCallback    UsageCallback
	functions        map[string]Function
	streamCallbackFn StreamCallback
	toolChoice       *string
	cache            *cache.Cache
	responseFormat   *ResponseFormat
	Name             string
}

// WithModel sets the model to use for the Anthropic instance.
func (a *Anthropic) WithModel(model Model) *Anthropic {
	a.model = model
	return a
}

// WithTemperature sets the temperature to use for the Anthropic instance.
func (a *Anthropic) WithTemperature(temperature float64) *Anthropic {
	a.temperature = temperature
	return a
}

// WithMaxTokens sets the max tokens to use for the Anthropic instance.
func (a *Anthropic) WithMaxTokens(maxTokens int) *Anthropic {
	a.maxTokens = maxTokens
	return a
}

// WithUsageCallback sets the usage callback to use for the Anthropic instance.
func (a *Anthropic) WithUsageCallback(callback UsageCallback) *Anthropic {
	a.usageCallback = callback
	return a
}

// WithStop sets the stop sequences to use for the Anthropic instance.
func (a *Anthropic) WithStop(stop []string) *Anthropic {
	a.stop = stop
	return a
}

// WithClient sets the client to use for the Anthropic instance.
func (a *Anthropic) WithClient(client *anthropicsdk.Client) *Anthropic {
	a.client = client
	return a
}

// WithToolChoice sets the tool choice to use for the Anthropic instance.
func (a *Anthropic) WithToolChoice(toolChoice *string) *Anthropic {
	a.toolChoice = toolChoice
	return a
}

// WithStream enables or disables streaming with a callback function.
func (a *Anthropic) WithStream(enable bool, callbackFn StreamCallback) *Anthropic {
	if !enable {
		a.streamCallbackFn = nil
	} else {
		a.streamCallbackFn = callbackFn
	}

	return a
}

// WithCache sets the cache to use for the Anthropic instance.
func (a *Anthropic) WithCache(cache *cache.Cache) *Anthropic {
	a.cache = cache
	return a
}

// WithResponseFormat sets the response format to use for the Anthropic instance.
// Use ResponseFormatJSONObject to get responses in JSON format.
// Use ResponseFormatText to get responses in plain text format (default).
func (a *Anthropic) WithResponseFormat(responseFormat ResponseFormat) *Anthropic {
	a.responseFormat = &responseFormat
	return a
}

// getCache retrieves a cached response for the given thread if available.
func (a *Anthropic) getCache(ctx context.Context, t *thread.Thread) (*cache.Result, error) {
	messages := t.UserQuery()
	cacheQuery := strings.Join(messages, "\n")
	cacheResult, err := a.cache.Get(ctx, cacheQuery)
	if err != nil {
		return cacheResult, err
	}

	t.AddMessage(thread.NewAssistantMessage().AddContent(
		thread.NewTextContent(strings.Join(cacheResult.Answer, "\n")),
	))

	return cacheResult, nil
}

// setCache stores the response in cache for future retrieval.
func (a *Anthropic) setCache(ctx context.Context, t *thread.Thread, cacheResult *cache.Result) error {
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

	err := a.cache.Set(ctx, cacheResult.Embedding, strings.Join(contents, "\n"))
	if err != nil {
		return err
	}

	return nil
}

// Add setUsageMetadata method for consistent usage reporting
func (a *Anthropic) setUsageMetadata(usage types.Meta) {
	if a.usageCallback == nil {
		return
	}

	callbackMetadata := make(types.Meta)
	for k, v := range usage {
		callbackMetadata[k] = v
	}

	a.usageCallback(callbackMetadata)
}

// New creates a new Anthropic instance with API key from environment
func New() *Anthropic {
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")

	return &Anthropic{
		client:      anthropicsdk.NewClient(option.WithAPIKey(anthropicKey)),
		model:       ModelClaude3_5SonnetLatest,
		temperature: DefaultAnthropicTemperature,
		maxTokens:   DefaultAnthropicMaxTokens,
		stop:        []string{},
		functions:   make(map[string]Function),
		Name:        "anthropic",
	}
}

// NewAnthropic creates a new Anthropic instance with explicit API key.
func NewAnthropic(apiKey string) *Anthropic {
	client := anthropicsdk.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Anthropic{
		client:      client,
		model:       ModelClaude3_5SonnetLatest,
		temperature: DefaultAnthropicTemperature,
		maxTokens:   DefaultAnthropicMaxTokens,
		stop:        []string{},
		functions:   make(map[string]Function),
		Name:        "anthropic",
	}
}

// stream handles streaming responses from the Anthropic API.
func (a *Anthropic) stream(ctx context.Context, t *thread.Thread, request anthropicsdk.MessageNewParams) error {
	stream := a.client.Messages.NewStreaming(ctx, request)
	var assistantMessage string
	var toolUses []anthropicsdk.ContentBlock
	var totalTokens int

	for stream.Next() {
		event := stream.Current()
		switch e := event.AsUnion().(type) {
		case anthropicsdk.ContentBlockDeltaEvent:
			if e.Delta.Type == deltaTypeText {
				assistantMessage += e.Delta.Text
				if a.streamCallbackFn != nil {
					a.streamCallbackFn(e.Delta.Text)
				}
				totalTokens++
			} else if e.Delta.Type == deltaTypeToolUse {
				toolUses = append(toolUses, anthropicsdk.ContentBlock{
					Type:  "tool_use",
					Input: json.RawMessage(e.Delta.Text),
				})
			}
		case anthropicsdk.MessageStopEvent:
			if a.streamCallbackFn != nil {
				a.streamCallbackFn(EOS)
			}

			// Process usage if callback is provided
			if a.usageCallback != nil {
				// Estimate token usage - Anthropic doesn't provide exact counts in streaming
				usageData := types.Meta{
					"output_tokens": totalTokens,
				}
				a.setUsageMetadata(usageData)
			}
		}
	}

	if stream.Err() != nil {
		return fmt.Errorf("%w: %s", ErrAnthropicChat, stream.Err())
	}

	var messages []*thread.Message
	if len(assistantMessage) > 0 {
		messages = append(messages, thread.NewAssistantMessage().AddContent(
			thread.NewTextContent(assistantMessage),
		))
	}

	if len(toolUses) > 0 {
		messages = append(messages, a.callTools(toolUses)...)
	}

	t.AddMessages(messages...)
	return nil
}

// generate handles non-streaming responses from the Anthropic API.
func (a *Anthropic) generate(ctx context.Context, t *thread.Thread, request anthropicsdk.MessageNewParams) error {
	response, err := a.client.Messages.New(ctx, request)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAnthropicChat, err)
	}

	// Process usage information if a callback is provided
	if a.usageCallback != nil {
		usageData := types.Meta{
			"input_tokens":  response.Usage.InputTokens,
			"output_tokens": response.Usage.OutputTokens,
		}
		a.setUsageMetadata(usageData)
	}

	var messages []*thread.Message
	var toolUses []anthropicsdk.ContentBlock

	for _, content := range response.Content {
		if content.Type == anthropicsdk.ContentBlockTypeToolUse {
			toolUses = append(toolUses, content)
		} else if content.Type == anthropicsdk.ContentBlockTypeText {
			messages = append(messages, thread.NewAssistantMessage().AddContent(
				thread.NewTextContent(content.Text),
			))
		}
	}

	if len(toolUses) > 0 {
		messages = append(messages, a.callTools(toolUses)...)
	}

	t.AddMessages(messages...)
	return nil
}

// buildChatCompletionRequest constructs the chat completion request parameters.
func (a *Anthropic) buildChatCompletionRequest(t *thread.Thread) anthropicsdk.MessageNewParams {
	messages := threadToChatCompletionMessages(t)
	messageParams := make([]anthropicsdk.MessageParam, len(messages))
	for i, msg := range messages {
		contentParams := make([]anthropicsdk.ContentBlockParamUnion, len(msg.Content))
		for j, content := range msg.Content {
			switch content.Type {
			case anthropicsdk.ContentBlockTypeText:
				contentParams[j] = anthropicsdk.ContentBlockParam{
					Type: anthropicsdk.F(anthropicsdk.ContentBlockParamTypeText),
					Text: anthropicsdk.F(content.Text),
				}
			case "image":
				contentParams[j] = anthropicsdk.ContentBlockParam{
					Type:   anthropicsdk.F(anthropicsdk.ContentBlockParamTypeImage),
					Source: anthropicsdk.F(interface{}(content.Input)),
				}
			}
		}

		messageParams[i] = anthropicsdk.MessageParam{
			Role:    anthropicsdk.F(anthropicsdk.MessageParamRole(msg.Role)),
			Content: anthropicsdk.F(contentParams),
		}
	}

	var toolChoice anthropicsdk.ToolChoiceUnionParam
	if a.toolChoice == nil || *a.toolChoice == "auto" {
		toolChoice = anthropicsdk.ToolChoiceAutoParam{
			Type: anthropicsdk.F(anthropicsdk.ToolChoiceAutoTypeAuto),
		}
	} else {
		toolChoice = anthropicsdk.ToolChoiceToolParam{
			Type: anthropicsdk.F(anthropicsdk.ToolChoiceToolTypeTool),
			Name: anthropicsdk.F(*a.toolChoice),
		}
	}

	params := anthropicsdk.MessageNewParams{
		Model:       anthropicsdk.F(anthropicsdk.Model(a.model)),
		Messages:    anthropicsdk.F(messageParams),
		MaxTokens:   anthropicsdk.F(int64(a.maxTokens)),
		Temperature: anthropicsdk.F(a.temperature),
		Tools:       anthropicsdk.F(a.getChatCompletionRequestTools()),
		ToolChoice:  anthropicsdk.F(toolChoice),
	}

	// Add stop sequences if provided
	if len(a.stop) > 0 {
		stopSequences := make([]string, len(a.stop))
		copy(stopSequences, a.stop)
		params.StopSequences = anthropicsdk.F(stopSequences)
	}

	return params
}

// getChatCompletionRequestTools retrieves the tools to include in the request.
func (a *Anthropic) getChatCompletionRequestTools() []anthropicsdk.ToolParam {
	var tools []anthropicsdk.ToolParam
	for _, function := range a.functions {
		tools = append(tools, anthropicsdk.ToolParam{
			Name:        anthropicsdk.F(function.Name),
			Description: anthropicsdk.F(function.Description),
			InputSchema: anthropicsdk.F(interface{}(function.Parameters)),
		})
	}
	return tools
}

// callTool executes the specified tool and returns the result as JSON.
func (a *Anthropic) callTool(toolUse anthropicsdk.ContentBlock) (string, error) {
	fn, ok := a.functions[string(toolUse.Type)]
	if !ok {
		return "", fmt.Errorf("unknown function %s", toolUse.Type)
	}

	resultAsJSON, err := callFnWithArgumentAsJSON(fn.Fn, string(toolUse.Input))
	if err != nil {
		return "", err
	}

	return resultAsJSON, nil
}

// callTools processes a list of tool uses and returns corresponding thread messages.
func (a *Anthropic) callTools(toolUses []anthropicsdk.ContentBlock) []*thread.Message {
	if len(a.functions) == 0 || len(toolUses) == 0 {
		return nil
	}

	var messages []*thread.Message
	for _, toolUse := range toolUses {
		result, err := a.callTool(toolUse)
		if err != nil {
			result = fmt.Sprintf("error: %s", err)
		}

		toolParam := anthropicsdk.ToolParam{
			Name:        anthropicsdk.F(string(toolUse.Type)),
			InputSchema: anthropicsdk.F(interface{}(string(toolUse.Input))),
		}
		messages = append(messages, toolCallResultToThreadMessage(toolParam, result))
	}

	return messages
}

// StartObserveGeneration initiates observation of message generation.
func (a *Anthropic) startObserveGeneration(ctx context.Context, t *thread.Thread) (*observer.Generation, error) {
	return llmobserver.StartObserveGeneration(
		ctx,
		a.Name,
		string(a.model),
		types.M{
			"maxTokens":   a.maxTokens,
			"temperature": a.temperature,
		},
		t,
	)
}

// StopObserveGeneration concludes observation of message generation.
func (a *Anthropic) stopObserveGeneration(
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

// Generate implements the LLM interface for generating responses.
func (a *Anthropic) Generate(ctx context.Context, t *thread.Thread) error {
	if t == nil {
		return nil
	}

	var err error
	var cacheResult *cache.Result
	if a.cache != nil {
		cacheResult, err = a.getCache(ctx, t)
		if err == nil {
			return nil
		} else if !errors.Is(err, cache.ErrCacheMiss) {
			return fmt.Errorf("%w: %w", ErrAnthropicChat, err)
		}
	}

	// If JSON format is requested, add a system message explicitly requesting JSON
	if a.responseFormat != nil && *a.responseFormat == ResponseFormatJSONObject {
		// Add a special system message requesting JSON format
		t.AddMessage(thread.NewSystemMessage().AddContent(
			thread.NewTextContent(`I need you to format your entire response as a valid JSON object. 
Your response should consist entirely of properly formatted JSON.`),
		))
	}

	generation, err := a.startObserveGeneration(ctx, t)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAnthropicChat, err)
	}

	// Store the message count before generation to track only new messages
	nMessageBeforeGeneration := len(t.Messages)

	request := a.buildChatCompletionRequest(t)

	// Process the request
	if a.streamCallbackFn != nil {
		err = a.stream(ctx, t, request)
	} else {
		err = a.generate(ctx, t, request)
	}

	if err != nil {
		// Wrap all errors with ErrAnthropicChat for consistent error handling
		return fmt.Errorf("%w: %w", ErrAnthropicChat, err)
	}

	err = a.stopObserveGeneration(ctx, generation, t.Messages[nMessageBeforeGeneration:])
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAnthropicChat, err)
	}

	if a.cache != nil {
		err = a.setCache(ctx, t, cacheResult)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrAnthropicChat, err)
		}
	}

	return nil
}

// Chat implements the LLM interface
func (a *Anthropic) Chat(ctx context.Context, t *thread.Thread) error {
	return a.Generate(ctx, t)
}

// WithFunctions implements the LLM interface
func (a *Anthropic) WithFunctions(functions map[string]Function) *Anthropic {
	a.functions = functions
	return a
}

// GetFunctions implements the LLM interface
func (a *Anthropic) GetFunctions() map[string]Function {
	return a.functions
}

// SetStop sets the stop sequences for the Anthropic instance.
func (a *Anthropic) SetStop(stop []string) {
	a.stop = stop
}
