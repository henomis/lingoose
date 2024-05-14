package assistant

import (
	"context"
	"strings"

	obs "github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

type Parameters struct {
	AssistantName      string
	AssistantIdentity  string
	AssistantScope     string
	CompanyName        string
	CompanyDescription string
}

type observer interface {
	Span(s *obs.Span) (*obs.Span, error)
	SpanEnd(s *obs.Span) (*obs.Span, error)
}

type Assistant struct {
	llm             LLM
	rag             RAG
	thread          *thread.Thread
	parameters      Parameters
	observer        observer
	observerTraceID string
}

type LLM interface {
	Generate(context.Context, *thread.Thread) error
}

type RAG interface {
	Retrieve(ctx context.Context, query string) ([]string, error)
}

func New(llm LLM) *Assistant {
	assistant := &Assistant{
		llm:    llm,
		thread: thread.New(),
		parameters: Parameters{
			AssistantName:      defaultAssistantName,
			AssistantIdentity:  defaultAssistantIdentity,
			AssistantScope:     defaultAssistantScope,
			CompanyName:        defaultCompanyName,
			CompanyDescription: defaultCompanyDescription,
		},
	}

	return assistant
}

func (a *Assistant) WithThread(thread *thread.Thread) *Assistant {
	a.thread = thread
	return a
}

func (a *Assistant) WithRAG(rag RAG) *Assistant {
	a.rag = rag
	return a
}

func (a *Assistant) WithParameters(parameters Parameters) *Assistant {
	a.parameters = parameters
	return a
}

func (a *Assistant) WithObserver(observer observer, traceID string) *Assistant {
	a.observer = observer
	a.observerTraceID = traceID
	return a
}

func (a *Assistant) Run(ctx context.Context) error {
	if a.thread == nil {
		return nil
	}

	var err error
	var spanAssistant *obs.Span
	if a.observer != nil {
		spanAssistant, err = a.startObserveSpan(ctx, "assistant")
		if err != nil {
			return err
		}
		ctx = obs.ContextWithParentID(ctx, spanAssistant.ID)
	}

	if a.rag != nil {
		err := a.generateRAGMessage(ctx)
		if err != nil {
			return err
		}
	}

	err = a.llm.Generate(ctx, a.thread)
	if err != nil {
		return err
	}

	if a.observer != nil {
		err = a.stopObserveSpan(spanAssistant)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Assistant) RunWithThread(ctx context.Context, thread *thread.Thread) error {
	a.thread = thread
	return a.Run(ctx)
}

func (a *Assistant) Thread() *thread.Thread {
	return a.thread
}

func (a *Assistant) generateRAGMessage(ctx context.Context) error {
	lastMessage := a.thread.LastMessage()
	if lastMessage.Role != thread.RoleUser || len(lastMessage.Contents) == 0 {
		return nil
	}

	query := strings.Join(a.thread.UserQuery(), "\n")
	a.thread.Messages = a.thread.Messages[:len(a.thread.Messages)-1]

	searchResults, err := a.rag.Retrieve(ctx, query)
	if err != nil {
		return err
	}

	a.thread.AddMessage(thread.NewSystemMessage().AddContent(
		thread.NewTextContent(
			systemRAGPrompt,
		).Format(
			types.M{
				"assistantName":      a.parameters.AssistantName,
				"assistantIdentity":  a.parameters.AssistantIdentity,
				"assistantScope":     a.parameters.AssistantScope,
				"companyName":        a.parameters.CompanyName,
				"companyDescription": a.parameters.CompanyDescription,
			},
		),
	)).AddMessage(thread.NewUserMessage().AddContent(
		thread.NewTextContent(
			baseRAGPrompt,
		).Format(
			types.M{
				"question": query,
				"results":  searchResults,
			},
		),
	))

	return nil
}

func (a *Assistant) startObserveSpan(ctx context.Context, name string) (*obs.Span, error) {
	return a.observer.Span(
		&obs.Span{
			TraceID:  a.observerTraceID,
			ParentID: obs.ContextValueParentID(ctx),
			Name:     name,
			Input:    a.parameters,
		},
	)
}

func (a *Assistant) stopObserveSpan(span *obs.Span) error {
	_, err := a.observer.SpanEnd(span)
	return err
}
