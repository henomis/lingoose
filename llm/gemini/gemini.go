package gemini

import (
	"cloud.google.com/go/vertexai/genai"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/henomis/lingoose/llm/cache"
	"github.com/henomis/lingoose/thread"
	"google.golang.org/api/iterator"
	"strings"
)

const (
	EOS = "\x00"
)

var threadRoleToGeminiRole = map[thread.Role]string{
	thread.RoleSystem:    "system_instruction",
	thread.RoleUser:      "user",
	thread.RoleAssistant: "model",
	thread.RoleTool:      "tool",
}

type Gemini struct {
	ctx              context.Context
	client           *genai.Client
	model            Model
	genModel         *genai.GenerativeModel
	session          *genai.ChatSession
	temperature      float32
	maxTokens        int
	stop             []string
	functions        map[string]Function
	streamCallbackFn StreamCallback
	tools            []*genai.Tool
	cache            *cache.Cache
	currentParts     []genai.Part
}

// WithTemperature sets the temperature to use for the Gemini instance.
func (g *Gemini) WithTemperature(temperature float32) *Gemini {
	g.temperature = temperature
	return g
}

func (g *Gemini) WithStream(enable bool, callbackFn StreamCallback) *Gemini {
	if !enable {
		g.streamCallbackFn = nil
	} else {
		g.streamCallbackFn = callbackFn
	}
	return g
}

func (g *Gemini) ClearTools() {
	g.genModel.Tools = []*genai.Tool{}
}

func (g *Gemini) WithTools(tools []*genai.Tool) *Gemini {
	// TODO: Handle genai.Schema from caller function
	g.genModel.Tools = tools
	return g
}

func (g *Gemini) WithCache(cache *cache.Cache) *Gemini {
	g.cache = cache
	return g
}

func (g *Gemini) WithChatMode() *Gemini {
	g.session = g.genModel.StartChat()
	return g
}

func New(ctx context.Context, client *genai.Client, model Model) *Gemini {
	gemini := &Gemini{}
	gemini.ctx = ctx
	gemini.model = model
	gemini.client = client
	gemini.functions = make(map[string]Function)
	gemini.genModel = gemini.client.GenerativeModel(model.String())
	return gemini
}

func (g *Gemini) GetChatHistory() []*genai.Content {
	if g.session != nil {
		return g.session.History
	}
	return nil
}

func (g *Gemini) GetTools() []*genai.Tool {
	return g.genModel.Tools
}

func (g *Gemini) GetTokenCount() (*genai.CountTokensResponse, error) {
	return g.genModel.CountTokens(g.ctx, g.currentParts...)
}

func (g *Gemini) getCache(ctx context.Context, t *thread.Thread) (*cache.Result, error) {
	messages := t.UserQuery()
	cacheQuery := strings.Join(messages, "\n")
	cacheResult, err := g.cache.Get(ctx, cacheQuery)
	if err != nil {
		return cacheResult, err
	}

	t.AddMessage(thread.NewAssistantMessage().AddContent(
		thread.NewTextContent(strings.Join(cacheResult.Answer, "\n")),
	))

	return cacheResult, nil
}

func (g *Gemini) setCache(ctx context.Context, t *thread.Thread, cacheResult *cache.Result) error {
	lastMessage := t.LastMessage()

	if lastMessage.Role != thread.RoleAssistant || len(lastMessage.Contents) == 0 {
		return nil
	}

	contents := make([]string, 0)
	for _, content := range lastMessage.Contents {
		if content.Type == thread.ContentTypeText {
			contents = append(contents, content.Data.(string))
		} else {
			contents = make([]string, 0)
			break
		}
	}

	err := g.cache.Set(ctx, cacheResult.Embedding, strings.Join(contents, "\n"))
	if err != nil {
		return err
	}

	return nil
}

func (g *Gemini) Generate(ctx context.Context, t *thread.Thread) error {
	if t == nil {
		return nil
	}
	var err error
	var cacheResult *cache.Result
	if g.cache != nil {
		cacheResult, err = g.getCache(ctx, t)
		if err == nil {
			return nil
		} else if !errors.Is(err, cache.ErrCacheMiss) {
			return fmt.Errorf("%w: %w", ErrGeminiChat, err)
		}
	}
	var (
		errChat error
		errGen  error
		parts   []genai.Part
	)
	defer func() {
		if len(parts) != 0 {
			g.currentParts = parts
		}
	}()

	if g.session != nil {
		parts, errChat = g.buildChatRequest(t)
		if errChat != nil {
			return err
		}
		if g.streamCallbackFn != nil {
			errChat = g.streamChat(ctx, t, parts)
		} else {
			errChat = g.generateChat(ctx, t, parts)
		}
		if errChat != nil {
			return errChat
		}
	} else {
		parts = g.buildRequest(t)
		if g.streamCallbackFn != nil {
			errGen = g.stream(ctx, t, parts)
		} else {
			errGen = g.generate(ctx, t, parts)
		}
		if errGen != nil {
			return errGen
		}
	}

	if g.cache != nil {
		err = g.setCache(ctx, t, cacheResult)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrGeminiChat, err)
		}
	}

	return nil
}

func (g *Gemini) stream(ctx context.Context, t *thread.Thread, parts []genai.Part) error {
	iter := g.genModel.GenerateContentStream(ctx, parts...)

	var (
		messages            []*thread.Message
		currentFuncToolCall genai.FunctionCall
		allFuncToolCall     []genai.FunctionCall
		content             strings.Builder
	)

	for {
		response, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			g.streamCallbackFn(EOS)
			if content.Len() > 0 {
				messages = append(messages, thread.NewAssistantMessage().AddContent(
					thread.NewTextContent(content.String()),
				))
			}

			if currentFuncToolCall.Name != "" {
				messages = append(messages, functionToolCallsToToolCallMessage(allFuncToolCall))
				messages = append(messages, g.callFuncTools(allFuncToolCall)...)
			}

			break
		}

		if response == nil || err != nil {
			return fmt.Errorf("%w", err)
		}

		if len(response.Candidates) == 0 {
			out, _ := json.Marshal(response.PromptFeedback)
			return fmt.Errorf("no candidates retured | prompt feedback: %s", string(out))
		}

		//check func tool call
		part := response.Candidates[0].Content.Parts[0]
		funCall, ok := part.(genai.FunctionCall)
		if ok {
			allFuncToolCall = append(allFuncToolCall, funCall)
			currentFuncToolCall = funCall
		} else {
			content.WriteString(PartsTostring(response.Candidates[0].Content.Parts))
			g.streamCallbackFn(PartsTostring(response.Candidates[0].Content.Parts))
		}
	}
	t.AddMessages(messages...)
	return nil
}

func (g *Gemini) generate(ctx context.Context, t *thread.Thread, parts []genai.Part) error {

	response, err := g.genModel.GenerateContent(ctx, parts...)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGeminiChat, err)
	}

	if len(response.Candidates) == 0 {
		out, _ := json.Marshal(response.PromptFeedback)
		return fmt.Errorf("no candidates retured | prompt feedback: %s", string(out))
	}

	var messages []*thread.Message

	//check func tool call
	part := response.Candidates[0].Content.Parts[0]
	funCall, ok := part.(genai.FunctionCall)
	if ok {
		messages = append(messages, functionToolCallsToToolCallMessage([]genai.FunctionCall{funCall}))
		messages = append(messages, g.callFuncTools([]genai.FunctionCall{funCall})...)
	} else {
		messages = []*thread.Message{
			thread.NewAssistantMessage().AddContent(
				thread.NewTextContent(PartsTostring(response.Candidates[0].Content.Parts)),
			),
		}
	}

	t.Messages = append(t.Messages, messages...)

	return nil
}

func (g *Gemini) buildRequest(t *thread.Thread) []genai.Part {
	return threadToPartMessage(t)
}

func (g *Gemini) buildChatRequest(t *thread.Thread) ([]genai.Part, error) {
	return threadToChatPartMessage(t)
}

func (g *Gemini) callFuncTools(toolCalls []genai.FunctionCall) []*thread.Message {
	if len(g.functions) == 0 || len(toolCalls) == 0 {
		return nil
	}

	var messages []*thread.Message
	for _, toolCall := range toolCalls {
		result, err := g.callTool(toolCall)
		if err != nil {
			result = fmt.Sprintf("error: %s", err)
		}

		messages = append(messages, toolCallResultToThreadMessage(toolCall, result))
	}

	return messages
}

func (g *Gemini) callTool(fnc genai.FunctionCall) (string, error) {
	fn, ok := g.functions[fnc.Name]
	if !ok {
		return "", fmt.Errorf("unknown function %s", fnc.Name)
	}

	jsonArgs, err := json.Marshal(fnc.Args)
	if err != nil {
		return "", fmt.Errorf("error in marshal: %w", err)
	}

	resultAsJSON, err := callFnWithArgumentAsJSON(fn.Fn, string(jsonArgs))
	if err != nil {
		return "", err
	}

	return resultAsJSON, nil
}
