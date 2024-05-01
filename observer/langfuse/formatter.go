package langfuse

import (
	"github.com/henomis/langfuse-go/model"
	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/thread"
)

func langfuseTraceToObserverTrace(l *model.Trace) *observer.Trace {
	return &observer.Trace{
		ID:   l.ID,
		Name: l.Name,
	}
}

func observerTraceToLangfuseTrace(t *observer.Trace) *model.Trace {
	return &model.Trace{
		ID:   t.ID,
		Name: t.Name,
	}
}

func langfuseSpanToObserverSpan(s *model.Span) *observer.Span {
	return &observer.Span{
		ID:       s.ID,
		TraceID:  s.TraceID,
		Name:     s.Name,
		ParentID: s.ParentObservationID,
	}
}

func observerSpanToLangfuseSpan(s *observer.Span) *model.Span {
	return &model.Span{
		ID:                  s.ID,
		TraceID:             s.TraceID,
		Name:                s.Name,
		ParentObservationID: s.ParentID,
	}
}

func threadMessagesToLangfuseMSlice(messages []*thread.Message) []model.M {
	if len(messages) == 0 {
		return nil
	}

	var mSlice []model.M
	for _, message := range messages {
		mSlice = append(mSlice, threadMessageToLangfuseM(message))
	}
	return mSlice
}

func threadMessageToLangfuseM(message *thread.Message) model.M {
	if message == nil {
		return nil
	}

	messageContent := ""
	role := message.Role
	for _, content := range message.Contents {
		if content.Type == thread.ContentTypeText {
			messageContent += content.AsString()
		}
	}
	return model.M{
		"role":    role,
		"content": messageContent,
	}
}

func observerGenerationToLangfuseGeneration(g *observer.Generation) *model.Generation {
	return &model.Generation{
		ID:                  g.ID,
		TraceID:             g.TraceID,
		Name:                g.Name,
		ParentObservationID: g.ParentID,
		Model:               g.Model,
		ModelParameters:     g.ModelParameters,
		Input:               threadMessagesToLangfuseMSlice(g.Input),
		Output:              threadMessageToLangfuseM(g.Output),
		Metadata:            g.Metadata,
	}
}

func observerEventToLangfuseEvent(e *observer.Event) *model.Event {
	return &model.Event{
		ID:                  e.ID,
		ParentObservationID: e.ParentID,
		TraceID:             e.TraceID,
		Name:                e.Name,
		Metadata:            e.Metadata,
	}
}

func observerScoreToLangfuseScore(s *observer.Score) *model.Score {
	return &model.Score{
		ID:      s.ID,
		TraceID: s.TraceID,
		Name:    s.Name,
		Value:   s.Value,
	}
}
