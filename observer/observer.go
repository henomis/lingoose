package observer

import (
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

type Trace struct {
	ID   string
	Name string
}

type Span struct {
	ID       string
	ParentID string
	TraceID  string
	Name     string
}

type Generation struct {
	ID              string
	ParentID        string
	TraceID         string
	Name            string
	Model           string
	ModelParameters types.M
	Input           []*thread.Message
	Output          *thread.Message
	Metadata        types.M
}

type Event struct {
	ID       string
	ParentID string
	TraceID  string
	Name     string
	Metadata types.M
}

type Score struct {
	ID      string
	TraceID string
	Name    string
	Value   float64
}
