package assistant

import (
	"context"

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

type Assistant struct {
	llm        LLM
	rag        RAG
	thread     *thread.Thread
	parameters Parameters
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

func (a *Assistant) Run(ctx context.Context) error {
	if a.thread == nil {
		return nil
	}

	if a.rag != nil {
		err := a.generateRAGMessage(ctx)
		if err != nil {
			return err
		}
	}

	return a.llm.Generate(ctx, a.thread)
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

	query := ""
	for _, content := range lastMessage.Contents {
		if content.Type == thread.ContentTypeText {
			query += content.Data.(string) + "\n"
		} else {
			continue
		}
	}

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
