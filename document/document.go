package document

import "github.com/henomis/lingoose/types"

type Document struct {
	Content  string     `json:"content"`
	Metadata types.Meta `json:"metadata"`
}
