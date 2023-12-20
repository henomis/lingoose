package thread

type Thread struct {
	Messages []*Message
}

type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
	// ContentTypeVideo ContentType = "video"
	// ContentTypeAudio ContentType = "audio"
	ContentTypeTool ContentType = "tool"
)

type Content struct {
	Type ContentType
	Data any
}

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role     Role
	Contents []*Content
}

type ToolData struct {
	ID     string
	Name   string
	Result string
}

type MediaData struct {
	Raw any
	URL *string
}

func NewTextContent(text string) *Content {
	return &Content{
		Type: ContentTypeText,
		Data: text,
	}
}

func NewImageContent(mediaData *MediaData) *Content {
	return &Content{
		Type: ContentTypeImage,
		Data: mediaData,
	}
}

// func NewVideoContent(mediaData *MediaData) *Content {
// 	return &Content{
// 		Type: ContentTypeVideo,
// 		Data: mediaData,
// 	}
// }

// func NewAudioContent(mediaData *MediaData) *Content {
// 	return &Content{
// 		Type: ContentTypeAudio,
// 		Data: mediaData,
// 	}
// }

func NewToolContent(toolData *ToolData) *Content {
	return &Content{
		Type: ContentTypeTool,
		Data: toolData,
	}
}

func (m *Message) AddContent(content *Content) {
	m.Contents = append(m.Contents, content)
}

func NewUserMessage() *Message {
	return &Message{
		Role: RoleUser,
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

func (t *Thread) AddMessage(message *Message) {
	t.Messages = append(t.Messages, message)
}

func NewThread() *Thread {
	return &Thread{}
}
