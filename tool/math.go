package tool

import (
	"context"
	"go/token"
	"go/types"
)

type math struct {
	name        string
	description string
}

func NewMath() *math {
	return &math{
		name:        "calculator",
		description: "useful for getting the result of a math expression. The input to this tool should be a valid mathematical expression that could be executed by a simple calculator.",
	}
}

func (m *math) Name() string {
	return m.name
}

func (m *math) Description() string {
	return m.description
}

func (m *math) Execute(ctx context.Context, input string) (string, error) {
	fs := token.NewFileSet()
	tv, err := types.Eval(fs, nil, token.NoPos, input)
	if err != nil {
		panic(err)
	}

	val := tv.Value.String()

	return val, nil
}
