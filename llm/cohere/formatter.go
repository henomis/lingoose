package cohere

import (
	"encoding/json"

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

func threadToChatMessages(t *thread.Thread) (string, []model.ChatMessage, []model.ToolResult) {
	var history []model.ChatMessage
	var toolResults []model.ToolResult
	var message string

	for i, m := range t.Messages {

		switch m.Role {
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

func extractToolResultFromThread(t *thread.Thread, toolCallData thread.ToolCallData, index, nToolCall int) *model.ToolResult {
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
