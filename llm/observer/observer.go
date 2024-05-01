package observer

import obs "github.com/henomis/lingoose/observer"

type LLMObserver interface {
	Span(*obs.Span) (*obs.Span, error)
	SpanEnd(*obs.Span) (*obs.Span, error)
	Generation(*obs.Generation) (*obs.Generation, error)
	GenerationEnd(*obs.Generation) (*obs.Generation, error)
}
