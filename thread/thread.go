package thread

type Thread struct {
	Messages []Message
}

type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
	ContentTypeTool  ContentType = "tool"
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
	Contents []Content
}

type ToolData struct {
	ID     string
	Name   string
	Result string
}
