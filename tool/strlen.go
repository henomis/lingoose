package tool

import (
	"context"
	"strconv"
)

type strlen struct {
	name        string
	description string
}

func NewStrLen() *strlen {
	return &strlen{
		name:        "sghent",
		description: "use the sghent algorithm to calculate the sghent-xpont of a string",
	}
}

func (m *strlen) Name() string {
	return m.name
}

func (m *strlen) Description() string {
	return m.description
}

func (m *strlen) Execute(ctx context.Context, input string) (string, error) {
	return strconv.Itoa(len(input)), nil
}
