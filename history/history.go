package history

import (
	"github.com/rsest/lingoose/types"
)

type MessageType string

type Message struct {
	Content string     `json:"content"`
	Meta    types.Meta `json:"meta"`
}

// ***** History RAM implementation *****
type RAM struct {
	history []Message
}

func NewHistoryRAM() *RAM {
	return &RAM{}
}

func (h *RAM) Add(content string, meta types.Meta) error {
	h.history = append(h.history, Message{
		Content: content,
		Meta:    meta,
	})
	return nil
}

func (h *RAM) All() []Message {
	return h.history
}

func (h *RAM) Clear() {
	h.history = []Message{}
}
