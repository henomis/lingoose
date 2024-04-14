package anthropic

import (
	"github.com/henomis/lingoose/thread"
	"strings"
)

func (o *Anthropic) buildChatCompletionRequest(t *thread.Thread) *request {
	messages, systemPrompt := threadToChatMessages(t)

	return &request{
		Model:       string(o.model),
		Messages:    messages,
		Tools:       o.getToolsRequest(),
		System:      systemPrompt,
		MaxTokens:   o.maxTokens,
		Temperature: o.temperature,
	}
}

func (o *Anthropic) getToolsRequest() []tool {
	var tools []tool
	for _, fn := range o.functions {
		tools = append(tools, tool{
			Name:        fn.Name,
			Description: fn.Description,
			InputSchema: fn.InputSchema,
		})
	}
	return tools
}

//nolint:gocognit
func threadToChatMessages(t *thread.Thread) ([]message, string) {
	var systemPrompt string
	var chatMessages []message
	for _, m := range t.Messages {
		switch m.Role {
		case thread.RoleSystem:
			for _, content := range m.Contents {
				contentData, ok := content.Data.(string)
				if !ok {
					continue
				}

				systemPrompt += contentData
			}
		case thread.RoleUser, thread.RoleAssistant:
			chatMessage := message{
				Role: threadRoleToAnthropicRole[m.Role],
			}
			var appendToPreviousMessage bool
			if chatMessages != nil && chatMessages[len(chatMessages)-1].Role == chatMessage.Role {
				chatMessage = chatMessages[len(chatMessages)-1]
				appendToPreviousMessage = true
			}
			for _, c := range m.Contents {
				contentData, _ := c.Data.(string)

				switch c.Type {
				case thread.ContentTypeText:
					chatMessage.Content = append(
						chatMessage.Content,
						content{
							Type: messageTypeText,
							Text: &contentData,
						},
					)
				case thread.ContentTypeImage:
					imageData, mimeType, err := getImageDataAsBase64(contentData)
					if err != nil {
						continue
					}

					chatMessage.Content = append(
						chatMessage.Content,
						content{
							Type: messageTypeImage,
							Source: &contentSource{
								Type:      "base64",
								Data:      imageData,
								MediaType: mimeType,
							},
						},
					)
				case thread.ContentTypeToolCall:
					if data, isToolCallData := c.Data.([]thread.ToolCallData); isToolCallData {
						chatMessage.Content = append(
							chatMessage.Content,
							content{
								Type:  messageTypeToolUse,
								Id:    data[0].ID,
								Name:  data[0].Name,
								Input: []byte(data[0].Arguments),
							},
						)
					}
				case thread.ContentTypeToolResponse:
					if data, isToolResponseData := c.Data.(thread.ToolResponseData); isToolResponseData {
						var isError bool
						if strings.Contains(data.Result, "Error: ") {
							isError = true
						}
						chatMessage.Content = append(
							chatMessage.Content,
							content{
								Type:      messageTypeToolResult,
								ToolUseId: data.ID,
								Content:   []byte(data.Result),
								IsError:   isError,
							},
						)
					}
				}
			}
			if !appendToPreviousMessage {
				chatMessages = append(chatMessages, chatMessage)
			} else {
				chatMessages[len(chatMessages)-1] = chatMessage
			}
		case thread.RoleTool:
			continue
		}
	}

	return chatMessages, systemPrompt
}

func toolCallsToToolCallContent(toolCall content) *thread.Content {
	var toolCallData []thread.ToolCallData
	toolCallData = append(toolCallData, thread.ToolCallData{
		ID:        toolCall.Id,
		Name:      toolCall.Name,
		Arguments: string(toolCall.Input),
	})

	return thread.NewToolCallContent(
		toolCallData,
	)
}
