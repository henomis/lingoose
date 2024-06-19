package cohere

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/henomis/cohere-go/model"
	"github.com/henomis/cohere-go/request"
	"github.com/henomis/lingoose/thread"
)

var threadRoleToCohereRole = map[thread.Role]model.ChatMessageRole{
	thread.RoleSystem:    model.ChatMessageRoleChatbot,
	thread.RoleUser:      model.ChatMessageRoleUser,
	thread.RoleAssistant: model.ChatMessageRoleChatbot,
	thread.RoleTool:      model.ChatMessageRoleTool,
}

func (c *Cohere) buildChatCompletionRequest(t *thread.Thread) *request.Chat {
	message, history, toolResults := threadToChatMessages(t)

	return &request.Chat{
		Model:         c.model,
		ChatHistory:   history,
		Message:       message,
		ToolResults:   toolResults,
		Temperature:   &c.temperature,
		MaxTokens:     &c.maxTokens,
		StopSequences: c.stop,
	}
}

// nolint:gocognit
func threadToChatMessages(t *thread.Thread) (string, []model.ChatMessage, []model.ToolResult) {
	var history []model.ChatMessage
	var toolResults []model.ToolResult
	var message string

	for i, m := range t.Messages {
		switch m.Role {
		case thread.RoleTool:
			//TODO: INSERT HERE TOOL RESULTS
			continue
		case thread.RoleUser, thread.RoleSystem:
			chatMessage := model.ChatMessage{
				Role: threadRoleToCohereRole[m.Role],
			}

			for _, content := range m.Contents {
				if content.Type == thread.ContentTypeText {
					chatMessage.Message += content.Data.(string) + "\n"
				}
			}

			history = append(history, chatMessage)

		case thread.RoleAssistant:
			// INSERT HERE TOOL CALLS
			// check if the message is a tool call
			if data, isTollCallData := m.Contents[0].Data.([]thread.ToolCallData); isTollCallData {
				toolResults = append(toolResults, threadToChatMessagesTool(data, t, i)...)
				continue
			}

			chatMessage := model.ChatMessage{
				Role: threadRoleToCohereRole[m.Role],
			}

			for _, content := range m.Contents {
				if content.Type == thread.ContentTypeText {
					chatMessage.Message += content.Data.(string) + "\n"
				}
			}

			history = append(history, chatMessage)
		}
	}

	lastMessage := t.LastMessage()
	if lastMessage.Role == thread.RoleUser {
		for _, content := range lastMessage.Contents {
			if content.Type == thread.ContentTypeText {
				message += content.Data.(string) + "\n"
			}
		}

		history = history[:len(history)-1]
	}

	return message, history, toolResults
}

func threadToChatMessagesTool(data []thread.ToolCallData, t *thread.Thread, index int) []model.ToolResult {
	var toolResults []model.ToolResult

	for nToolCall, toolCallData := range data {
		toolResult := extractToolResultFromThread(t, toolCallData, index, nToolCall)
		if toolResult != nil {
			toolResults = append(toolResults, *toolResult)
		}
	}

	return toolResults
}

func extractToolResultFromThread(
	t *thread.Thread,
	toolCallData thread.ToolCallData,
	index,
	nToolCall int,
) *model.ToolResult {
	messageIndex := index + 1 + nToolCall

	// check if the message index is within the bounds of the thread
	if messageIndex >= len(t.Messages) {
		return nil
	}

	message := t.Messages[messageIndex]

	// check if the message role is a tool
	if message.Role != thread.RoleTool {
		return nil
	}

	if data, isTollResponseData := message.Contents[0].Data.(thread.ToolResponseData); isTollResponseData {
		argumentsAsMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(toolCallData.Arguments), &argumentsAsMap)
		if err != nil {
			return nil
		}

		resultsAsMap := make(map[string]interface{})
		err = json.Unmarshal([]byte(data.Result), &resultsAsMap)
		if err != nil {
			return nil
		}

		return &model.ToolResult{
			Call: model.ToolCall{
				Name:       toolCallData.Name,
				Parameters: argumentsAsMap,
			},
			Outputs: []interface{}{
				resultsAsMap,
			},
		}
	}

	return nil
}

func buildChatCompletionRequestTool(function Function) *model.Tool {
	tool := model.Tool{
		Name:                 function.Name,
		Description:          function.Description,
		ParameterDefinitions: make(map[string]model.ToolParameterDefinition),
	}

	functionProperties, ok := function.Parameters["properties"]
	if !ok {
		return nil
	}

	functionPropertiesAsMap, isMap := functionProperties.(map[string]interface{})
	if !isMap {
		return nil
	}

	for k, v := range functionPropertiesAsMap {
		toolParameterDefinitions := buildFunctionParameterDefinitions(v)
		if toolParameterDefinitions != nil {
			tool.ParameterDefinitions[k] = *toolParameterDefinitions
		}
	}

	return &tool
}

func buildFunctionParameterDefinitions(v any) *model.ToolParameterDefinition {
	valueAsMap, isValueMap := v.(map[string]interface{})
	if !isValueMap {
		return nil
	}

	description := ""
	descriptionValue, isDescriptionValue := valueAsMap["description"]
	if isDescriptionValue {
		descriptionAsString, isDescriptionString := descriptionValue.(string)
		if isDescriptionString {
			description = descriptionAsString
		}
	}

	argType := ""
	argTypeValue, isArgTypeValue := valueAsMap["type"]
	if isArgTypeValue {
		argTypeAsString, isArgTypeString := argTypeValue.(string)
		if isArgTypeString {
			argType = argTypeAsString
		}
	}

	required := false
	requiredValue, isRequiredValue := valueAsMap["required"]
	if isRequiredValue {
		requiredAsBool, isRequiredBool := requiredValue.(bool)
		if isRequiredBool {
			required = requiredAsBool
		}
	}

	if description == "" || argType == "" {
		return nil
	}

	return &model.ToolParameterDefinition{
		Description: description,
		Type:        argType,
		Required:    required,
	}
}

func toolCallsToToolCallMessage(toolCalls []model.ToolCall) *thread.Message {
	if len(toolCalls) == 0 {
		return nil
	}

	var toolCallData []thread.ToolCallData
	for i, toolCall := range toolCalls {
		parametersAsString, err := json.Marshal(toolCall.Parameters)
		if err != nil {
			continue
		}

		if toolCalls[i].ID == "" {
			toolCalls[i].ID = uuid.New().String()
		}

		toolCallData = append(toolCallData, thread.ToolCallData{
			ID:        toolCalls[i].ID,
			Name:      toolCall.Name,
			Arguments: string(parametersAsString),
		})
	}

	return thread.NewAssistantMessage().AddContent(
		thread.NewToolCallContent(
			toolCallData,
		),
	)
}

func toolCallResultToThreadMessage(toolCall model.ToolCall, result string) *thread.Message {
	return thread.NewToolMessage().AddContent(
		thread.NewToolResponseContent(
			thread.ToolResponseData{
				ID:     toolCall.ID,
				Name:   toolCall.Name,
				Result: result,
			},
		),
	)
}
