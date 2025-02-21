package thread

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/rsest/lingoose/types"
)

type Thread struct {
	Messages []*Message
}

type ContentType string

const (
	ContentTypeText         ContentType = "text"
	ContentTypeImage        ContentType = "image"
	ContentTypeToolCall     ContentType = "tool_call"
	ContentTypeToolResponse ContentType = "tool_response"
)

type Content struct {
	Type ContentType
	Data any
}

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role     Role
	Contents []*Content
}

type ToolResponseData struct {
	ID     string
	Name   string
	Result string
}

type ToolCallData struct {
	ID        string
	Name      string
	Arguments string
}

func NewTextContent(text string) *Content {
	return &Content{
		Type: ContentTypeText,
		Data: text,
	}
}

func NewImageContentFromURL(url string) *Content {
	return &Content{
		Type: ContentTypeImage,
		Data: url,
	}
}

func NewToolResponseContent(toolResponseData ToolResponseData) *Content {
	return &Content{
		Type: ContentTypeToolResponse,
		Data: toolResponseData,
	}
}

func NewToolCallContent(data []ToolCallData) *Content {
	return &Content{
		Type: ContentTypeToolCall,
		Data: data,
	}
}

func (m *Message) AddContent(content *Content) *Message {
	m.Contents = append(m.Contents, content)
	return m
}

func NewUserMessage() *Message {
	return &Message{
		Role: RoleUser,
	}
}

func NewSystemMessage() *Message {
	return &Message{
		Role: RoleSystem,
	}
}

func NewAssistantMessage() *Message {
	return &Message{
		Role: RoleAssistant,
	}
}

func NewToolMessage() *Message {
	return &Message{
		Role: RoleTool,
	}
}

func (t *Thread) AddMessage(message *Message) *Thread {
	t.Messages = append(t.Messages, message)
	return t
}

func (t *Thread) AddMessages(messages ...*Message) *Thread {
	t.Messages = append(t.Messages, messages...)
	return t
}

func (t *Thread) CountMessages() int {
	return len(t.Messages)
}

func New() *Thread {
	return &Thread{}
}

func (t *Thread) String() string {
	str := "Thread:\n"
	for _, message := range t.Messages {
		str += string(message.Role) + ":\n"
		for _, content := range message.Contents {
			str += "\tType: " + string(content.Type) + "\n"
			switch content.Type {
			case ContentTypeText:
				str += "\tText: " + content.Data.(string) + "\n"
			case ContentTypeImage:
				if contentAsString, ok := content.Data.(string); ok {
					str += "\tImage URL: " + contentAsString + "\n"
				}
			case ContentTypeToolCall:
				for _, toolCallData := range content.Data.([]ToolCallData) {
					str += "\tTool Call ID: " + toolCallData.ID + "\n"
					str += "\tTool Call Function Name: " + toolCallData.Name + "\n"
					str += "\tTool Call Function Arguments: " + toolCallData.Arguments + "\n"
				}
			case ContentTypeToolResponse:
				str += "\tTool ID: " + content.Data.(ToolResponseData).ID + "\n"
				str += "\tTool Name: " + content.Data.(ToolResponseData).Name + "\n"
				str += "\tTool Result: " + content.Data.(ToolResponseData).Result + "\n"
			}
		}
	}
	return str
}

// LastMessage returns the last message in the thread.
func (t *Thread) LastMessage() *Message {
	return t.Messages[len(t.Messages)-1]
}

// UserQuery returns the last user messages as a slice of strings.
func (t *Thread) UserQuery() []string {
	userMessages := make([]*Message, 0)
	for _, message := range t.Messages {
		if message.Role == RoleUser {
			userMessages = append(userMessages, message)
		} else {
			userMessages = make([]*Message, 0)
		}
	}

	var messages []string
	for _, message := range userMessages {
		for _, content := range message.Contents {
			if content.Type == ContentTypeText {
				messages = append(messages, content.Data.(string))
			} else {
				messages = make([]string, 0)
				break
			}
		}
	}

	return messages
}

func (t *Thread) ClearMessages() *Thread {
	t.Messages = make([]*Message, 0)
	return t
}

func (m *Message) ClearContents() *Message {
	m.Contents = make([]*Content, 0)
	return m
}

func (c *Content) Format(input types.M) *Content {
	if c.Type != ContentTypeText || input == nil {
		return c
	}

	if !strings.Contains(c.Data.(string), "{{") {
		return c
	}

	templateEngine, err := template.New("prompt").
		Option("missingkey=zero").Parse(c.Data.(string))
	if err != nil {
		return c
	}

	var buffer bytes.Buffer
	err = templateEngine.Execute(&buffer, input)
	if err != nil {
		return c
	}

	c.Data = buffer.String()

	return c
}

func (c *Content) AsString() string {
	if contentAsString, ok := c.Data.(string); ok {
		return contentAsString
	}
	return ""
}

func (c *Content) AsToolResponseData() *ToolResponseData {
	if contentAsToolResponseData, ok := c.Data.(ToolResponseData); ok {
		return &contentAsToolResponseData
	}
	return nil
}

func (c *Content) AsToolCallData() []ToolCallData {
	if contentAsToolCallData, ok := c.Data.([]ToolCallData); ok {
		return contentAsToolCallData
	}
	return nil
}
