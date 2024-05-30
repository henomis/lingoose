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

const (
	DefaultMaxIterations = 3
)

type Assistant struct {
	llm           LLM
	rag           RAG
	thread        *thread.Thread
	parameters    Parameters
	maxIterations uint
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
		maxIterations: DefaultMaxIterations,
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

func (a *Assistant) Run(ctx context.Context) error {
	if a.thread == nil {
		return nil
	}

	ctx, spanAssistant, err := a.startObserveSpan(ctx, "assistant")
	if err != nil {
		return err
	}

	if a.rag != nil {
		errGenerate := a.generateRAGMessage(ctx)
		if errGenerate != nil {
			return errGenerate
		}
	}

	err = a.llm.Generate(ctx, a.thread)
	if err != nil {
		return err
	}

	err = a.stopObserveSpan(ctx, spanAssistant)
	if err != nil {
		return err
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
			systemPrompt,
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

func (a *Assistant) WithMaxIterations(maxIterations uint) *Assistant {
	a.maxIterations = maxIterations
	return a
}

func (a *Assistant) Execute(ctx context.Context) error {
	if a.thread == nil {
		return nil
	}

	ctx, spanAssistant, err := a.startObserveSpan(ctx, "assistant")
	if err != nil {
		return err
	}

	a.injectSystemMessage()

	for i := 0; i < int(a.maxIterations); i++ {
		err = a.llm.Generate(ctx, a.thread)
		if err != nil {
			return err
		}

		if a.thread.LastMessage().Role != thread.RoleTool {
			break
		}
	}

	err = a.stopObserveSpan(ctx, spanAssistant)
	if err != nil {
		return err
	}

	return nil
}

func (a *Assistant) startObserveSpan(ctx context.Context, name string) (context.Context, *obs.Span, error) {
	o, ok := obs.ContextValueObserverInstance(ctx).(observer)
	if o == nil || !ok {
		// No observer instance in context
		return ctx, nil, nil
	}

	span, err := o.Span(
		&obs.Span{
			TraceID:  obs.ContextValueTraceID(ctx),
			ParentID: obs.ContextValueParentID(ctx),
			Name:     name,
			Input:    a.parameters,
		},
	)
	if err != nil {
		return ctx, nil, err
	}

	if span != nil {
		ctx = obs.ContextWithParentID(ctx, span.ID)
	}

	return ctx, span, nil
}

func (a *Assistant) stopObserveSpan(ctx context.Context, span *obs.Span) error {
	o, ok := obs.ContextValueObserverInstance(ctx).(observer)
	if o == nil || !ok {
		// No observer instance in context
		return nil
	}

	_, err := o.SpanEnd(span)
	return err
}

func (a *Assistant) injectSystemMessage() {
	for _, message := range a.thread.Messages {
		if message.Role == thread.RoleSystem {
			return
		}
	}

	systemMessage := thread.NewSystemMessage().AddContent(
		thread.NewTextContent(
			systemPrompt,
		).Format(
			types.M{
				"assistantName":      a.parameters.AssistantName,
				"assistantIdentity":  a.parameters.AssistantIdentity,
				"assistantScope":     a.parameters.AssistantScope,
				"companyName":        a.parameters.CompanyName,
				"companyDescription": a.parameters.CompanyDescription,
			},
		),
	)

	a.thread.Messages = append([]*thread.Message{systemMessage}, a.thread.Messages...)
}
