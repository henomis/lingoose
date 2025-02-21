// Package chat provides a chat prompt template.
// Sometimes you need to define a chat prompt, this package provides a way to do that.
package chat

import (
	"errors"
	"fmt"

	"github.com/rsest/lingoose/types"
)

var (
	ErrChatMessages = errors.New("unable to convert chat messages")
)

type Chat struct {
	promptMessages PromptMessages
}

type Prompt interface {
	String() string
	Format(input types.M) error
}

type MessageType string

const (
	MessageTypeSystem    MessageType = "system"
	MessageTypeUser      MessageType = "user"
	MessageTypeAssistant MessageType = "assistant"
	MessageTypeFunction  MessageType = "function"
)

type PromptMessage struct {
	Type   MessageType
	Prompt Prompt
	Name   *string
}

type PromptMessages []PromptMessage

type Message struct {
	Type    MessageType
	Content string
	Name    *string
}

type Messages []Message

func New(promptMessages ...PromptMessage) *Chat {
	chatPromptTemplate := &Chat{}
	for _, message := range promptMessages {
		chatPromptTemplate.addMessagePromptTemplate(message)
	}

	return chatPromptTemplate
}

// AddPromptMessages adds a list of chat prompt templates to the chat prompt template.
func (c *Chat) AddPromptMessages(messages []PromptMessage) {
	for _, message := range messages {
		c.addMessagePromptTemplate(message)
	}
}

func (c *Chat) addMessagePromptTemplate(message PromptMessage) {
	c.promptMessages = append(c.promptMessages, message)
}

// ToMessages converts the chat prompt template to a list of messages.
func (c *Chat) ToMessages() (Messages, error) {
	var messages Messages
	var err error

	for _, messagePromptTemplate := range c.promptMessages {
		var message Message
		message.Type = messagePromptTemplate.Type
		message.Name = messagePromptTemplate.Name

		if messagePromptTemplate.Prompt != nil {
			if len(messagePromptTemplate.Prompt.String()) == 0 {
				err = messagePromptTemplate.Prompt.Format(types.M{})
				if err != nil {
					return nil, fmt.Errorf("%w: %w", ErrChatMessages, err)
				}
			}
			message.Content = messagePromptTemplate.Prompt.String()
		}

		messages = append(messages, message)
	}

	return messages, nil
}

// PromptMessages returns the chat prompt messages.
func (c *Chat) PromptMessages() PromptMessages {
	return c.promptMessages
}
