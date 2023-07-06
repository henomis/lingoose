package history

import (
	"github.com/henomis/lingoose/types"
)

type HistoryMessageType string

type HistoryMessage struct {
	Content string     `json:"content"`
	Meta    types.Meta `json:"meta"`
}

// ***** History RAM implementation *****
type HistoryRam struct {
	history []HistoryMessage
}

func NewHistoryRam() *HistoryRam {
	return &HistoryRam{}
}

func (h *HistoryRam) Add(content string, meta types.Meta) error {
	h.history = append(h.history, HistoryMessage{
		Content: content,
		Meta:    meta,
	})
	return nil
}

func (h *HistoryRam) All() []HistoryMessage {
	return h.history
}

func (h *HistoryRam) Clear() {
	h.history = []HistoryMessage{}
}
