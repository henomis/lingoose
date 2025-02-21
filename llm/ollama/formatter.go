package ollama

import (
	"github.com/rsest/lingoose/thread"
)

func (o *Ollama) buildChatCompletionRequest(t *thread.Thread) *request {
	return &request{
		Model:    o.model,
		Messages: threadToChatMessages(t),
		Options: options{
			Temperature: o.temperature,
		},
	}
}

//nolint:gocognit
func threadToChatMessages(t *thread.Thread) []message {
	var chatMessages []message
	for _, m := range t.Messages {
		switch m.Role {
		case thread.RoleUser, thread.RoleSystem, thread.RoleAssistant:
			for _, content := range m.Contents {
				chatMessage := message{
					Role: threadRoleToOllamaRole[m.Role],
				}

				contentData, ok := content.Data.(string)
				if !ok {
					continue
				}

				if content.Type == thread.ContentTypeText {
					chatMessage.Content = contentData
				} else if content.Type == thread.ContentTypeImage {
					imageData, err := getImageDataAsBase64(contentData)
					if err != nil {
						continue
					}
					chatMessage.Images = []string{imageData}
				} else {
					continue
				}

				chatMessages = append(chatMessages, chatMessage)
			}
		case thread.RoleTool:
			continue
		}
	}

	return chatMessages
}
