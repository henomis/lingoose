package anthropic

import (
	"encoding/json"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/henomis/lingoose/thread"
)

//nolint:gocognit
func threadToChatCompletionMessages(t *thread.Thread) []anthropicsdk.Message {
	chatCompletionMessages := make([]anthropicsdk.Message, len(t.Messages))
	for i, message := range t.Messages {
		chatCompletionMessages[i] = anthropicsdk.Message{
			Role: anthropicsdk.MessageRole(message.Role),
		}

		// Create content blocks for this message
		var contentBlocks []anthropicsdk.ContentBlock

		// Process all contents in the message
		for _, content := range message.Contents {
			switch content.Type {
			case thread.ContentTypeText:
				if textData, ok := content.Data.(string); ok {
					contentBlocks = append(contentBlocks, anthropicsdk.ContentBlock{
						Type: anthropicsdk.ContentBlockTypeText,
						Text: textData,
					})
				}
			case thread.ContentTypeImage:
				// For image content, we need to handle it specially
				// The URL can be stored directly as a string
				if urlStr, ok := content.Data.(string); ok {
					// Create a custom image block with the proper structure
					imageContentBlock := anthropicsdk.ContentBlock{
						Type: "image",
					}

					// Create a JSON-based source element for the image
					sourceJSON := map[string]interface{}{
						"type": "url",
						"url":  urlStr,
					}
					sourceBytes, _ := json.Marshal(sourceJSON)

					// Set it directly into the internal field that the SDK uses
					err := json.Unmarshal(sourceBytes, &imageContentBlock.Input)
					if err != nil {
						panic(err)
					}

					contentBlocks = append(contentBlocks, imageContentBlock)
				}
			}
		}

		// Only set content if we have actual blocks
		if len(contentBlocks) > 0 {
			chatCompletionMessages[i].Content = contentBlocks
			continue
		}

		// Special handling for tool responses and calls
		if len(message.Contents) == 1 {
			switch message.Role {
			case thread.RoleAssistant:
				if toolCallDataSlice, isToolCallData := message.Contents[0].Data.([]thread.ToolCallData); isToolCallData {
					var toolUses []anthropicsdk.ContentBlock
					for _, toolCallData := range toolCallDataSlice {
						toolUses = append(toolUses, anthropicsdk.ContentBlock{
							Type:  "tool_use",
							Input: json.RawMessage(toolCallData.Arguments),
						})
					}
					chatCompletionMessages[i].Content = toolUses
				}
			case thread.RoleTool:
				if data, isToolResponseData := message.Contents[0].Data.(thread.ToolResponseData); isToolResponseData {
					chatCompletionMessages[i].Content = []anthropicsdk.ContentBlock{
						{
							Type: anthropicsdk.ContentBlockTypeText,
							Text: data.Result,
						},
					}
				}
			}
		}
	}

	return chatCompletionMessages
}

func toolCallResultToThreadMessage(toolCall anthropicsdk.ToolParam, result string) *thread.Message {
	return thread.NewToolMessage().AddContent(
		thread.NewToolResponseContent(
			thread.ToolResponseData{
				Name:   toolCall.Name.Value,
				Result: result,
			},
		),
	)
}
