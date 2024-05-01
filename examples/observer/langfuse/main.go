package main

import (
	"context"

	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/observer/langfuse"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

func main() {
	l := langfuse.New(context.Background())

	trace, err := l.Trace(
		&observer.Trace{
			Name: "trace",
		},
	)
	if err != nil {
		panic(err)
	}

	span, err := l.Span(
		&observer.Span{
			Name:    "span",
			TraceID: trace.ID,
		},
	)
	if err != nil {
		panic(err)
	}

	generation, err := l.Generation(
		&observer.Generation{
			ParentID: span.ID,
			TraceID:  trace.ID,
			Name:     "generation",
			Model:    "gpt-3.5-turbo",
			ModelParameters: types.M{
				"maxTokens":   "1000",
				"temperature": "0.9",
			},
			Input: []*thread.Message{
				{
					Role: thread.RoleSystem,
					Contents: []*thread.Content{
						{
							Type: thread.ContentTypeText,
							Data: "You are a helpful assistant.",
						},
					},
				},
				{
					Role: thread.RoleUser,
					Contents: []*thread.Content{
						{
							Type: thread.ContentTypeText,
							Data: "Please generate a summary of the following documents \nThe engineering department defined the following OKR goals...\nThe marketing department defined the following OKR goals...",
						},
					},
				},
			},
			Metadata: types.M{
				"key": "value",
			},
		},
	)
	if err != nil {
		panic(err)
	}

	_, err = l.Event(
		&observer.Event{
			ParentID: generation.ID,
			TraceID:  trace.ID,
			Name:     "event",
			Metadata: types.M{
				"key": "value",
			},
		},
	)
	if err != nil {
		panic(err)
	}

	generation.Output = &thread.Message{
		Role: thread.RoleAssistant,
		Contents: []*thread.Content{
			{
				Type: thread.ContentTypeText,
				Data: "The Q3 OKRs contain goals for multiple teams...",
			},
		},
	}

	_, err = l.GenerationEnd(generation)
	if err != nil {
		panic(err)
	}

	_, err = l.Score(
		&observer.Score{
			TraceID: trace.ID,
			Name:    "score",
			Value:   0.9,
		},
	)
	if err != nil {
		panic(err)
	}

	_, err = l.SpanEnd(span)
	if err != nil {
		panic(err)
	}

	l.Flush(context.Background())
}
