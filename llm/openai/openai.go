package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/henomis/lingoose/llm/cache"
	llmobserver "github.com/henomis/lingoose/llm/observer"
	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
	"github.com/mitchellh/mapstructure"
	openai "github.com/sashabaranov/go-openai"
)

const (
	EOS = "\x00"
)

var threadRoleToOpenAIRole = map[thread.Role]string{
	thread.RoleSystem:    "system",
	thread.RoleUser:      "user",
	thread.RoleAssistant: "assistant",
	thread.RoleTool:      "tool",
}

type OpenAI struct {
	openAIClient     *openai.Client
	model            Model
	temperature      float32
	maxTokens        int
	stop             []string
	usageCallback    UsageCallback
	functions        map[string]Function
	streamCallbackFn StreamCallback
	responseFormat   *ResponseFormat
	toolChoice       *string
	cache            *cache.Cache
	Name             string
}

// WithModel sets the model to use for the OpenAI instance.
func (o *OpenAI) WithModel(model Model) *OpenAI {
	o.model = model
	return o
}

// WithTemperature sets the temperature to use for the OpenAI instance.
func (o *OpenAI) WithTemperature(temperature float32) *OpenAI {
	o.temperature = temperature
	return o
}

// WithMaxTokens sets the max tokens to use for the OpenAI instance.
func (o *OpenAI) WithMaxTokens(maxTokens int) *OpenAI {
	o.maxTokens = maxTokens
	return o
}

// WithUsageCallback sets the usage callback to use for the OpenAI instance.
func (o *OpenAI) WithUsageCallback(callback UsageCallback) *OpenAI {
	o.usageCallback = callback
	return o
}

// WithStop sets the stop sequences to use for the OpenAI instance.
func (o *OpenAI) WithStop(stop []string) *OpenAI {
	o.stop = stop
	return o
}

// WithClient sets the client to use for the OpenAI instance.
func (o *OpenAI) WithClient(client *openai.Client) *OpenAI {
	o.openAIClient = client
	return o
}

func (o *OpenAI) WithToolChoice(toolChoice *string) *OpenAI {
	o.toolChoice = toolChoice
	return o
}

func (o *OpenAI) WithStream(enable bool, callbackFn StreamCallback) *OpenAI {
	if !enable {
		o.streamCallbackFn = nil
	} else {
		o.streamCallbackFn = callbackFn
	}

	return o
}

func (o *OpenAI) WithCache(cache *cache.Cache) *OpenAI {
	o.cache = cache
	return o
}

func (o *OpenAI) WithResponseFormat(responseFormat ResponseFormat) *OpenAI {
	o.responseFormat = &responseFormat
	return o
}

// SetStop sets the stop sequences for the completion.
func (o *OpenAI) SetStop(stop []string) {
	o.stop = stop
}

func (o *OpenAI) setUsageMetadata(usage openai.Usage) {
	callbackMetadata := make(types.Meta)

	err := mapstructure.Decode(usage, &callbackMetadata)
	if err != nil {
		return
	}

	o.usageCallback(callbackMetadata)
}

func New() *OpenAI {
	openAIKey := os.Getenv("OPENAI_API_KEY")

	return &OpenAI{
		openAIClient: openai.NewClient(openAIKey),
		model:        GPT3Dot5Turbo,
		temperature:  DefaultOpenAITemperature,
		maxTokens:    DefaultOpenAIMaxTokens,
		functions:    make(map[string]Function),
		Name:         "openai",
	}
}

func (o *OpenAI) getCache(ctx context.Context, t *thread.Thread) (*cache.Result, error) {
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

func (o *OpenAI) setCache(ctx context.Context, t *thread.Thread, cacheResult *cache.Result) error {
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

func (o *OpenAI) Generate(ctx context.Context, t *thread.Thread) error {
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
			return fmt.Errorf("%w: %w", ErrOpenAIChat, err)
		}
	}

	chatCompletionRequest := o.BuildChatCompletionRequest(t)

	if len(o.functions) > 0 {
		chatCompletionRequest.Tools = o.GetChatCompletionRequestTools()
		chatCompletionRequest.ToolChoice = o.getChatCompletionRequestToolChoice()
	}

	generation, err := o.startObserveGeneration(ctx, t)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenAIChat, err)
	}

	if o.streamCallbackFn != nil {
		chatCompletionRequest.StreamOptions = &openai.StreamOptions{IncludeUsage: true}
		err = o.stream(ctx, t, chatCompletionRequest)
	} else {
		err = o.generate(ctx, t, chatCompletionRequest)
	}
	if err != nil {
		return err
	}

	err = o.stopObserveGeneration(ctx, generation, t)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenAIChat, err)
	}

	if o.cache != nil {
		err = o.setCache(ctx, t, cacheResult)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrOpenAIChat, err)
		}
	}

	return nil
}

func isStreamToolCallResponse(response *openai.ChatCompletionStreamResponse) bool {
	return response.Choices[0].FinishReason == openai.FinishReasonToolCalls || len(response.Choices[0].Delta.ToolCalls) > 0
}

func handleStreamToolCallResponse(
	response *openai.ChatCompletionStreamResponse,
	currentToolCall *openai.ToolCall,
) (updatedToolCall openai.ToolCall, isNewTool bool) {
	if len(response.Choices[0].Delta.ToolCalls) == 0 {
		return
	}
	if response.Choices[0].Delta.ToolCalls[0].ID != "" {
		isNewTool = true
		updatedToolCall = response.Choices[0].Delta.ToolCalls[0]
	} else {
		currentToolCall.Function.Arguments += response.Choices[0].Delta.ToolCalls[0].Function.Arguments
	}
	return
}

func (o *OpenAI) handleEndOfStream(
	messages []*thread.Message,
	content string,
	currentToolCall *openai.ToolCall,
	allToolCalls []openai.ToolCall,
) []*thread.Message {
	o.streamCallbackFn(EOS)
	if len(content) > 0 {
		messages = append(messages, thread.NewAssistantMessage().AddContent(
			thread.NewTextContent(content),
		))
	}
	if currentToolCall.ID != "" {
		allToolCalls = o.convertParallelToolCallsToIndividualCalls(append(allToolCalls, *currentToolCall))
		messages = append(messages, toolCallsToToolCallMessage(allToolCalls))
		messages = append(messages, o.callTools(allToolCalls)...)
	}
	return messages
}

func (o *OpenAI) stream(
	ctx context.Context,
	t *thread.Thread,
	chatCompletionRequest openai.ChatCompletionRequest,
) error {
	stream, err := o.openAIClient.CreateChatCompletionStream(
		ctx,
		chatCompletionRequest,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenAIChat, err)
	}

	var content string
	var messages []*thread.Message
	var allToolCalls []openai.ToolCall
	var currentToolCall openai.ToolCall
	for {
		response, errRecv := stream.Recv()
		if errors.Is(errRecv, io.EOF) {
			messages = o.handleEndOfStream(messages, content, &currentToolCall, allToolCalls)
			break
		}
		fmt.Printf("%+v\n", response)

		if len(response.Choices) == 0 {
			if response.Usage != nil {
				if o.usageCallback != nil {
					o.setUsageMetadata(*response.Usage)
				}
			} else {
				return fmt.Errorf("%w: no choices returned", ErrOpenAIChat)
			}
			continue
		}

		if isStreamToolCallResponse(&response) {
			updatedToolCall, isNewTool := handleStreamToolCallResponse(&response, &currentToolCall)
			if isNewTool {
				if currentToolCall.ID != "" {
					allToolCalls = append(allToolCalls, currentToolCall)
				}
				currentToolCall = updatedToolCall
			}
		} else {
			content += response.Choices[0].Delta.Content
		}

		o.streamCallbackFn(response.Choices[0].Delta.Content)
	}

	t.AddMessages(messages...)

	return nil
}

func (o *OpenAI) convertParallelToolCallsToIndividualCalls(toolCalls []openai.ToolCall) []openai.ToolCall {
	if len(toolCalls) == 0 {
		return nil
	}

	var toolCallData []openai.ToolCall
	for _, toolCall := range toolCalls {
		if toolCall.Function.Name == "multi_tool_use.parallel" {
			fmt.Println("Got parallel tool use", toolCall)
			var parallelToolCall struct {
				ToolUses []struct {
					RecipientName string `json:"recipient_name"`
					Parameters    string `json:"parameters"`
				} `json:"tool_uses"`
			}
			_ = json.Unmarshal([]byte(toolCall.Function.Arguments), &parallelToolCall)
			for index, toolUse := range parallelToolCall.ToolUses {
				toolCallData = append(toolCallData, openai.ToolCall{
					Index: &index,
					ID:    toolCall.ID,
					Type:  openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      strings.TrimPrefix("function.", toolUse.RecipientName),
						Arguments: toolUse.Parameters,
					},
				})
			}
		} else {
			toolCallData = append(toolCallData, toolCall)
		}
	}
	return toolCallData
}

func (o *OpenAI) generate(
	ctx context.Context,
	t *thread.Thread,
	chatCompletionRequest openai.ChatCompletionRequest,
) error {
	response, err := o.openAIClient.CreateChatCompletion(
		ctx,
		chatCompletionRequest,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenAIChat, err)
	}

	if o.usageCallback != nil {
		o.setUsageMetadata(response.Usage)
	}

	if len(response.Choices) == 0 {
		return fmt.Errorf("%w: no choices returned", ErrOpenAIChat)
	}

	var messages []*thread.Message
	if response.Choices[0].FinishReason == "tool_calls" || len(response.Choices[0].Message.ToolCalls) > 0 {
		toolCalls := o.convertParallelToolCallsToIndividualCalls(response.Choices[0].Message.ToolCalls)
		messages = append(messages, toolCallsToToolCallMessage(toolCalls))
		messages = append(messages, o.callTools(toolCalls)...)
	} else {
		messages = []*thread.Message{
			thread.NewAssistantMessage().AddContent(
				thread.NewTextContent(response.Choices[0].Message.Content),
			),
		}
	}

	t.Messages = append(t.Messages, messages...)

	return nil
}

func (o *OpenAI) BuildChatCompletionRequest(t *thread.Thread) openai.ChatCompletionRequest {
	var responseFormat *openai.ChatCompletionResponseFormat
	if o.responseFormat != nil {
		responseFormat = &openai.ChatCompletionResponseFormat{
			Type: *o.responseFormat,
		}
	}

	return openai.ChatCompletionRequest{
		Model:          string(o.model),
		Messages:       threadToChatCompletionMessages(t),
		MaxTokens:      o.maxTokens,
		Temperature:    o.temperature,
		N:              DefaultOpenAINumResults,
		TopP:           DefaultOpenAITopP,
		Stop:           o.stop,
		ResponseFormat: responseFormat,
	}
}

func (o *OpenAI) GetChatCompletionRequestTools() []openai.Tool {
	tools := []openai.Tool{}

	for _, function := range o.functions {
		tools = append(tools, openai.Tool{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        function.Name,
				Description: function.Description,
				Parameters:  function.Parameters,
			},
		})
	}

	return tools
}

func (o *OpenAI) getChatCompletionRequestToolChoice() any {
	if o.toolChoice == nil {
		return "none"
	}

	if *o.toolChoice == "auto" {
		return "auto"
	}

	return openai.ToolChoice{
		Type: openai.ToolTypeFunction,
		Function: openai.ToolFunction{
			Name: *o.toolChoice,
		},
	}
}

func (o *OpenAI) callTool(toolCall openai.ToolCall) (string, error) {
	fn, ok := o.functions[toolCall.Function.Name]
	if !ok {
		return "", fmt.Errorf("unknown function %s", toolCall.Function.Name)
	}

	resultAsJSON, err := callFnWithArgumentAsJSON(fn.Fn, toolCall.Function.Arguments)
	if err != nil {
		return "", err
	}

	return resultAsJSON, nil
}

func (o *OpenAI) callTools(toolCalls []openai.ToolCall) []*thread.Message {
	if len(o.functions) == 0 || len(toolCalls) == 0 {
		return nil
	}

	var messages []*thread.Message
	for _, toolCall := range toolCalls {
		result, err := o.callTool(toolCall)
		if err != nil {
			result = fmt.Sprintf("error: %s", err)
		}

		messages = append(messages, toolCallResultToThreadMessage(toolCall, result))
	}

	return messages
}

func (o *OpenAI) startObserveGeneration(ctx context.Context, t *thread.Thread) (*observer.Generation, error) {
	return llmobserver.StartObserveGeneration(
		ctx,
		o.Name,
		string(o.model),
		types.M{
			"maxTokens":   o.maxTokens,
			"temperature": o.temperature,
		},
		t,
	)
}

func (o *OpenAI) stopObserveGeneration(
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
