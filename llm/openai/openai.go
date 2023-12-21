package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/henomis/lingoose/thread"
	openai "github.com/sashabaranov/go-openai"
)

const (
	EOS = "\x00"
)

var threadRoleToOpenAIRole = map[thread.Role]string{
	thread.RoleUser:      "user",
	thread.RoleAssistant: "assistant",
	thread.RoleTool:      "tool",
}

func New() *OpenAI {
	openAIKey := os.Getenv("OPENAI_API_KEY")

	return &OpenAI{
		openAIClient:           openai.NewClient(openAIKey),
		model:                  GPT3Dot5Turbo,
		temperature:            DefaultOpenAITemperature,
		maxTokens:              DefaultOpenAIMaxTokens,
		verbose:                false,
		functions:              make(map[string]Function),
		functionsMaxIterations: DefaultMaxIterations,
	}
}

func (o *OpenAI) Stream(ctx context.Context, t *thread.Thread, callbackFn StreamCallback) error {
	if t == nil {
		return nil
	}

	chatCompletionRequest := o.buildChatCompletionRequest(t)

	if len(o.functions) > 0 {
		chatCompletionRequest.Tools = o.getChatCompletionRequestTools()
		chatCompletionRequest.ToolChoice = o.getChatCompletionRequestToolChoice()
	}

	stream, err := o.openAIClient.CreateChatCompletionStream(
		ctx,
		chatCompletionRequest,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenAIChat, err)
	}

	var messages []*thread.Message
	var content string
	for {
		response, errRecv := stream.Recv()
		if errors.Is(errRecv, io.EOF) {
			callbackFn(EOS)

			if len(content) > 0 {
				messages = append(messages, thread.NewAssistantMessage().AddContent(
					thread.NewTextContent(content),
				))
			}
			break
		}

		if len(response.Choices) == 0 {
			return fmt.Errorf("%w: no choices returned", ErrOpenAIChat)
		}

		if response.Choices[0].FinishReason == "tool_calls" || len(response.Choices[0].Delta.ToolCalls) > 0 {
			messages = append(messages, o.callTools(response.Choices[0].Delta.ToolCalls)...)
		} else {
			content += response.Choices[0].Delta.Content
		}

		callbackFn(response.Choices[0].Delta.Content)
	}

	t.AddMessages(messages)

	return nil
}

func (o *OpenAI) Generate(ctx context.Context, t *thread.Thread) error {
	if t == nil {
		return nil
	}

	chatCompletionRequest := o.buildChatCompletionRequest(t)

	if len(o.functions) > 0 {
		chatCompletionRequest.Tools = o.getChatCompletionRequestTools()
		chatCompletionRequest.ToolChoice = o.getChatCompletionRequestToolChoice()
	}

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
		messages = append(messages, toolCallsToToolCallMessage(response.Choices[0].Message.ToolCalls))
		messages = append(messages, o.callTools(response.Choices[0].Message.ToolCalls)...)
	} else {
		messages = []*thread.Message{
			textContentToThreadMessage(response.Choices[0].Message.Content),
		}
	}

	t.Messages = append(t.Messages, messages...)

	return nil
}

func (o *OpenAI) buildChatCompletionRequest(t *thread.Thread) openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model:       string(o.model),
		Messages:    threadToChatCompletionMessages(t),
		MaxTokens:   o.maxTokens,
		Temperature: o.temperature,
		N:           DefaultOpenAINumResults,
		TopP:        DefaultOpenAITopP,
		Stop:        o.stop,
	}
}

func (o *OpenAI) getChatCompletionRequestTools() []openai.Tool {
	tools := []openai.Tool{}

	for _, function := range o.functions {
		tools = append(tools, openai.Tool{
			Type: "function",
			Function: openai.FunctionDefinition{
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

	o.calledFunctionName = &fn.Name

	return resultAsJSON, nil
}

func (o *OpenAI) callTools(toolCalls []openai.ToolCall) []*thread.Message {
	if len(o.functions) == 0 || len(toolCalls) == 0 {
		return nil
	}

	var messages []*thread.Message
	for _, toolCall := range toolCalls {
		if o.verbose {
			fmt.Printf("Calling function %s\n", toolCall.Function.Name)
			fmt.Printf("Function call arguments: %s\n", toolCall.Function.Arguments)
		}

		result, err := o.callTool(toolCall)
		if err != nil {
			result = fmt.Sprintf("error: %s", err)
		}

		messages = append(messages, toolCallResultToThreadMessage(toolCall, result))
	}

	return messages
}
