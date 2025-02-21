package langfuse

import (
	"github.com/henomis/langfuse-go/model"
	"github.com/rsest/lingoose/observer"
	"github.com/rsest/lingoose/thread"
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
		Input:    s.Input,
		Output:   s.Output,
	}
}

func observerSpanToLangfuseSpan(s *observer.Span) *model.Span {
	return &model.Span{
		ID:                  s.ID,
		TraceID:             s.TraceID,
		Name:                s.Name,
		ParentObservationID: s.ParentID,
		Input:               s.Input,
		Output:              s.Output,
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

func threadOutputMessagesToLangfuseOutput(messages []*thread.Message) any {
	if len(messages) == 1 &&
		messages[0].Role == thread.RoleAssistant &&
		len(messages[0].Contents) == 1 &&
		messages[0].Contents[0].Type == thread.ContentTypeText {
		return threadMessageToLangfuseM(messages[0])
	}

	toolCalls := model.M{}
	toolMessages := []*thread.Message{}

	for _, message := range messages {
		if message.Role == thread.RoleAssistant &&
			message.Contents[0].Type == thread.ContentTypeToolCall {
			toolCalls = threadMessageToLangfuseM(message)
		} else if message.Role == thread.RoleTool &&
			message.Contents[0].Type == thread.ContentTypeToolResponse {
			toolMessages = append(toolMessages, message)
		}
	}

	return append([]model.M{toolCalls}, threadMessagesToLangfuseMSlice(toolMessages)...)
}

func threadMessageToLangfuseM(message *thread.Message) model.M {
	if message == nil {
		return nil
	}

	role := message.Role
	if message.Role == thread.RoleTool {
		data := message.Contents[0].AsToolResponseData()
		m := model.M{
			"type":    message.Contents[0].Type,
			"id":      data.ID,
			"name":    data.Name,
			"results": data.Result,
		}

		return model.M{
			"role":    role,
			"content": m,
		}
	}

	messageContent := ""
	m := make([]model.M, 0)
	for _, content := range message.Contents {
		if content.Type == thread.ContentTypeText {
			messageContent += content.AsString()
		} else if content.Type == thread.ContentTypeToolCall {
			for _, data := range content.AsToolCallData() {
				m = append(m, model.M{
					"type":      content.Type,
					"id":        data.ID,
					"name":      data.Name,
					"arguments": data.Arguments,
				})
			}
		}
	}
	output := model.M{
		"role":    role,
		"content": messageContent,
	}

	if len(m) > 0 {
		output["content"] = m
	}

	return output
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
		Output:              threadOutputMessagesToLangfuseOutput(g.Output),
		Metadata:            g.Metadata,
	}
}

func observerEmbeddingToLangfuseGeneration(e *observer.Embedding) *model.Generation {
	return &model.Generation{
		ID:                  e.ID,
		TraceID:             e.TraceID,
		Name:                e.Name,
		ParentObservationID: e.ParentID,
		Model:               e.Model,
		ModelParameters:     e.ModelParameters,
		Input:               e.Input,
		Output:              e.Output,
		Metadata:            e.Metadata,
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
