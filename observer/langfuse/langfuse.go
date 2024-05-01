package langfuse

import (
	"context"
	"time"

	langfusego "github.com/henomis/langfuse-go"
	"github.com/henomis/lingoose/observer"
)

type Langfuse struct {
	client *langfusego.Langfuse
}

func New(ctx context.Context) *Langfuse {
	return &Langfuse{
		client: langfusego.New(ctx),
	}
}

func (l *Langfuse) WithFlushInterval(d time.Duration) *Langfuse {
	l.client = l.client.WithFlushInterval(d)
	return l
}

func (l *Langfuse) Trace(t *observer.Trace) (*observer.Trace, error) {
	langfuseTrace := observerTraceToLangfuseTrace(t)
	langfuseTrace, err := l.client.Trace(langfuseTrace)
	if err != nil {
		return nil, err
	}
	return langfuseTraceToObserverTrace(langfuseTrace), nil
}

func (l *Langfuse) Span(s *observer.Span) (*observer.Span, error) {
	langfuseSpan := observerSpanToLangfuseSpan(s)
	langfuseSpan, err := l.client.Span(langfuseSpan, nil)
	if err != nil {
		return nil, err
	}
	return langfuseSpanToObserverSpan(langfuseSpan), nil
}

func (l *Langfuse) SpanEnd(s *observer.Span) (*observer.Span, error) {
	langfuseSpan := observerSpanToLangfuseSpan(s)
	_, err := l.client.SpanEnd(langfuseSpan)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (l *Langfuse) Generation(g *observer.Generation) (*observer.Generation, error) {
	langfuseGeneration := observerGenerationToLangfuseGeneration(g)
	langfuseGeneration, err := l.client.Generation(langfuseGeneration, nil)
	if err != nil {
		return nil, err
	}
	g.ID = langfuseGeneration.ID
	return g, nil
}

func (l *Langfuse) GenerationEnd(g *observer.Generation) (*observer.Generation, error) {
	langfuseGeneration := observerGenerationToLangfuseGeneration(g)
	_, err := l.client.Generation(langfuseGeneration, nil)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (l *Langfuse) Event(e *observer.Event) (*observer.Event, error) {
	langfuseEvent := observerEventToLangfuseEvent(e)
	langfuseEvent, err := l.client.Event(langfuseEvent, nil)
	if err != nil {
		return nil, err
	}
	e.ID = langfuseEvent.ID
	return e, nil
}

func (l *Langfuse) Score(s *observer.Score) (*observer.Score, error) {
	langfuseScore := observerScoreToLangfuseScore(s)
	langfuseScore, err := l.client.Score(langfuseScore)
	if err != nil {
		return nil, err
	}
	s.ID = langfuseScore.ID
	return s, nil
}

func (l *Langfuse) Flush(ctx context.Context) {
	l.client.Flush(ctx)
}
