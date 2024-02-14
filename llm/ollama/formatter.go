package ollama

import (
	"context"
	"fmt"

	"github.com/henomis/lingoose/thread"
)

const visionPrompt = "What is in this picture?"
const contentImageAsText = "The original content was an image, however the image content has been converted by an AI to text as an image description and can be used as context.\n\nImage description: %s"

func (o *Ollama) buildChatCompletionRequest(ctx context.Context, t *thread.Thread) *request {
	return &request{
		Model:    o.model,
		Messages: o.threadToChatMessages(ctx, t),
		Options: options{
			Temperature: o.temperature,
		},
	}
}

func (o *Ollama) buildVisionRequest(imageURL string) (*visionRequest, error) {
	imageData, err := getImageDataAsBase64(imageURL)
	if err != nil {
		return nil, err
	}

	return &visionRequest{
		Model:  *o.visionModel,
		Prompt: visionPrompt,
		Images: []string{imageData},
	}, nil
}

func (o *Ollama) threadToChatMessages(ctx context.Context, t *thread.Thread) []message {
	var chatMessages []message
	for messageIndex, m := range t.Messages {
		chatMessage := message{
			Role: threadRoleToOllamaRole[m.Role],
		}

		switch m.Role {
		case thread.RoleUser, thread.RoleSystem, thread.RoleAssistant:
			for contentIndex, content := range m.Contents {
				if content.Type == thread.ContentTypeText {
					chatMessage.Content += content.Data.(string) + "\n"
					chatMessages = append(chatMessages, chatMessage)
				} else if content.Type == thread.ContentTypeImage {
					visionRequest, err := o.buildVisionRequest(content.Data.(string))
					if err != nil {
						continue
					}

					content, err := o.vision(ctx, visionRequest)
					if err != nil {
						continue
					}
					chatMessage.Role = threadRoleToOllamaRole[thread.RoleAssistant]
					chatMessage.Content = fmt.Sprintf(contentImageAsText, *content)
					chatMessages = append(chatMessages, chatMessage)

					if o.convertImageContentToText {
						t.Messages[messageIndex].Contents[contentIndex].Type = thread.ContentTypeText
						t.Messages[messageIndex].Contents[contentIndex].Data = chatMessage.Content
					}
				}
			}
		case thread.RoleTool:
			continue
		}
	}

	return chatMessages
}
