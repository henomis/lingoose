package antropic

import (
	"github.com/henomis/lingoose/thread"
)

func (o *Antropic) buildChatCompletionRequest(t *thread.Thread) *request {
	messages, systemPrompt := threadToChatMessages(t)

	return &request{
		Model:       o.model,
		Messages:    messages,
		System:      systemPrompt,
		MaxTokens:   o.maxTokens,
		Temperature: o.temperature,
	}
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
				Role: threadRoleToOllamaRole[m.Role],
			}
			for _, c := range m.Contents {
				contentData, ok := c.Data.(string)
				if !ok {
					continue
				}

				if c.Type == thread.ContentTypeText {
					chatMessage.Content = append(
						chatMessage.Content,
						content{
							Type: messageTypeText,
							Text: &contentData,
						},
					)
				} else if c.Type == thread.ContentTypeImage {
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
				} else {
					continue
				}
			}
			chatMessages = append(chatMessages, chatMessage)
		case thread.RoleTool:
			continue
		}
	}

	return chatMessages, systemPrompt
}
