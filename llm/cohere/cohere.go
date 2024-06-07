package cohere

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	coherego "github.com/henomis/cohere-go"
	"github.com/henomis/cohere-go/model"
	"github.com/henomis/cohere-go/request"
	"github.com/henomis/cohere-go/response"

	"github.com/henomis/lingoose/legacy/chat"
	"github.com/henomis/lingoose/llm/cache"
	llmobserver "github.com/henomis/lingoose/llm/observer"
	"github.com/henomis/lingoose/observer"
	"github.com/henomis/lingoose/thread"
	"github.com/henomis/lingoose/types"
)

var (
	ErrCohereChat = fmt.Errorf("cohere chat error")
)

type Model = model.Model

const (
	ModelCommand             Model = model.ModelCommand
	ModelCommandNightly      Model = model.ModelCommandNightly
	ModelCommandLight        Model = model.ModelCommandLight
	ModelCommandLightNightly Model = model.ModelCommandLightNightly
	ModelCommandR            Model = model.ModelCommandR
	ModelCommandRPlus        Model = model.ModelCommandRPlus
)

const (
	DefaultMaxTokens   = 256
	DefaultTemperature = 0.75
	DefaultModel       = ModelCommand
)

type StreamCallbackFn func(string)

type Cohere struct {
	client           *coherego.Client
	model            Model
	temperature      float64
	maxTokens        int
	verbose          bool
	stop             []string
	cache            *cache.Cache
	streamCallbackFn StreamCallbackFn
	name             string
	functions        map[string]Function
}

func (c *Cohere) WithCache(cache *cache.Cache) *Cohere {
	c.cache = cache
	return c
}

// NewCompletion returns a new completion LLM
func NewCompletion() *Cohere {
	return New()
}

func New() *Cohere {
	return &Cohere{
		client:      coherego.New(os.Getenv("COHERE_API_KEY")),
		model:       DefaultModel,
		temperature: DefaultTemperature,
		maxTokens:   DefaultMaxTokens,
		name:        "cohere",
		functions:   make(map[string]Function),
	}
}

// WithModel sets the model to use for the LLM
func (c *Cohere) WithModel(model Model) *Cohere {
	c.model = model
	return c
}

// WithTemperature sets the temperature to use for the LLM
func (c *Cohere) WithTemperature(temperature float64) *Cohere {
	c.temperature = temperature
	return c
}

// WithMaxTokens sets the max tokens to use for the LLM
func (c *Cohere) WithMaxTokens(maxTokens int) *Cohere {
	c.maxTokens = maxTokens
	return c
}

// WithAPIKey sets the API key to use for the LLM
func (c *Cohere) WithAPIKey(apiKey string) *Cohere {
	c.client = coherego.New(apiKey)
	return c
}

// WithVerbose sets the verbosity of the LLM
func (c *Cohere) WithVerbose(verbose bool) *Cohere {
	c.verbose = verbose
	return c
}

// WithStop sets the stop sequences to use for the LLM
func (c *Cohere) WithStop(stop []string) *Cohere {
	c.stop = stop
	return c
}

func (c *Cohere) WithStream(callbackFn StreamCallbackFn) *Cohere {
	c.streamCallbackFn = callbackFn
	return c
}

// Completion returns the completion for the given prompt
func (c *Cohere) Completion(ctx context.Context, prompt string) (string, error) {
	resp := &response.Generate{}
	err := c.client.Generate(
		ctx,
		&request.Generate{
			Prompt:        prompt,
			Temperature:   &c.temperature,
			MaxTokens:     &c.maxTokens,
			Model:         &c.model,
			StopSequences: c.stop,
		},
		resp,
	)
	if err != nil {
		return "", err
	}

	if len(resp.Generations) == 0 {
		return "", fmt.Errorf("no generations returned")
	}

	output := resp.Generations[0].Text

	if c.verbose {
		fmt.Printf("---USER---\n%s\n", prompt)
		fmt.Printf("---AI---\n%s\n", output)
	}

	return output, nil
}

// Chat is not implemented
func (c *Cohere) Chat(ctx context.Context, prompt *chat.Chat) (string, error) {
	_ = ctx
	_ = prompt
	return "", fmt.Errorf("not implemented")
}

func (c *Cohere) getCache(ctx context.Context, t *thread.Thread) (*cache.Result, error) {
	messages := t.UserQuery()
	cacheQuery := strings.Join(messages, "\n")
	cacheResult, err := c.cache.Get(ctx, cacheQuery)
	if err != nil {
		return cacheResult, err
	}

	t.AddMessage(thread.NewAssistantMessage().AddContent(
		thread.NewTextContent(strings.Join(cacheResult.Answer, "\n")),
	))

	return cacheResult, nil
}

func (c *Cohere) setCache(ctx context.Context, t *thread.Thread, cacheResult *cache.Result) error {
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

	err := c.cache.Set(ctx, cacheResult.Embedding, strings.Join(contents, "\n"))
	if err != nil {
		return err
	}

	return nil
}

func (c *Cohere) Generate(ctx context.Context, t *thread.Thread) error {
	if t == nil {
		return nil
	}

	var err error
	var cacheResult *cache.Result
	if c.cache != nil {
		cacheResult, err = c.getCache(ctx, t)
		if err == nil {
			return nil
		} else if !errors.Is(err, cache.ErrCacheMiss) {
			return fmt.Errorf("%w: %w", ErrCohereChat, err)
		}
	}

	chatRequest := c.buildChatCompletionRequest(t)

	if len(c.functions) > 0 {
		chatRequest.Tools = c.getChatCompletionRequestTools()
	}

	generation, err := c.startObserveGeneration(ctx, t)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCohereChat, err)
	}

	nMessageBeforeGeneration := len(t.Messages)

	if c.streamCallbackFn != nil {
		err = c.stream(ctx, t, chatRequest)
	} else {
		err = c.generate(ctx, t, chatRequest)
	}
	if err != nil {
		return err
	}

	err = c.stopObserveGeneration(ctx, generation, t.Messages[nMessageBeforeGeneration:])
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCohereChat, err)
	}

	if c.cache != nil {
		err = c.setCache(ctx, t, cacheResult)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrCohereChat, err)
		}
	}

	return nil
}

func (c *Cohere) generate(ctx context.Context, t *thread.Thread, chatRequest *request.Chat) error {
	var response response.Chat

	err := c.client.Chat(
		ctx,
		chatRequest,
		&response,
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCohereChat, err)
	}

	var messages []*thread.Message
	if len(response.ToolCalls) > 0 {
		messages = append(messages, toolCallsToToolCallMessage(response.ToolCalls))
		messages = append(messages, c.callTools(response.ToolCalls)...)
	} else {
		messages = []*thread.Message{
			thread.NewAssistantMessage().AddContent(
				thread.NewTextContent(response.Text),
			),
		}
	}

	t.Messages = append(t.Messages, messages...)

	return nil
}

func (c *Cohere) stream(ctx context.Context, t *thread.Thread, chatRequest *request.Chat) error {
	chatResponse := &response.Chat{}
	var assistantMessage string

	err := c.client.ChatStream(
		ctx,
		chatRequest,
		chatResponse,
		func(r *response.Chat) {
			if r.Text != "" {
				c.streamCallbackFn(r.Text)
				assistantMessage += r.Text
			}
		},
	)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCohereChat, err)
	}

	t.AddMessage(thread.NewAssistantMessage().AddContent(
		thread.NewTextContent(assistantMessage),
	))

	return nil
}

func (c *Cohere) startObserveGeneration(ctx context.Context, t *thread.Thread) (*observer.Generation, error) {
	return llmobserver.StartObserveGeneration(
		ctx,
		c.name,
		string(c.model),
		types.M{
			"maxTokens":   c.maxTokens,
			"temperature": c.temperature,
		},
		t,
	)
}

func (c *Cohere) stopObserveGeneration(
	ctx context.Context,
	generation *observer.Generation,
	messages []*thread.Message,
) error {
	return llmobserver.StopObserveGeneration(
		ctx,
		generation,
		messages,
	)
}

func (o *Cohere) getChatCompletionRequestTools() []model.Tool {
	tools := []model.Tool{}

	for _, function := range o.functions {
		tool := model.Tool{
			Name:                 function.Name,
			Description:          function.Description,
			ParameterDefinitions: make(map[string]model.ToolParameterDefinition),
		}

		functionProperties, ok := function.Parameters["properties"]
		if !ok {
			continue
		}

		functionPropertiesAsMap, isMap := functionProperties.(map[string]interface{})
		if !isMap {
			continue
		}

		for k, v := range functionPropertiesAsMap {
			valueAsMap, isValueMap := v.(map[string]interface{})
			if !isValueMap {
				continue
			}

			description := ""
			descriptionValue, isDescriptionValue := valueAsMap["description"]
			if isDescriptionValue {
				descriptionAsString, isDescriptionString := descriptionValue.(string)
				if isDescriptionString {
					description = descriptionAsString
				}
			}

			argType := ""
			argTypeValue, isArgTypeValue := valueAsMap["type"]
			if isArgTypeValue {
				argTypeAsString, isArgTypeString := argTypeValue.(string)
				if isArgTypeString {
					argType = argTypeAsString
				}
			}

			required := false
			requiredValue, isRequiredValue := valueAsMap["required"]
			if isRequiredValue {
				requiredAsBool, isRequiredBool := requiredValue.(bool)
				if isRequiredBool {
					required = requiredAsBool
				}
			}

			if description == "" || argType == "" {
				continue
			}

			tool.ParameterDefinitions[k] = model.ToolParameterDefinition{
				Description: description,
				Type:        argType,
				Required:    required,
			}
		}

		tools = append(tools, tool)
	}

	return tools
}

func toolCallsToToolCallMessage(toolCalls []model.ToolCall) *thread.Message {
	if len(toolCalls) == 0 {
		return nil
	}

	var toolCallData []thread.ToolCallData
	for i, toolCall := range toolCalls {
		parametersAsString, err := json.Marshal(toolCall.Parameters)
		if err != nil {
			continue
		}

		if toolCalls[i].ID == "" {
			toolCalls[i].ID = uuid.New().String()
		}

		toolCallData = append(toolCallData, thread.ToolCallData{
			ID:        toolCalls[i].ID,
			Name:      toolCall.Name,
			Arguments: string(parametersAsString),
		})
	}

	return thread.NewAssistantMessage().AddContent(
		thread.NewToolCallContent(
			toolCallData,
		),
	)
}

func toolCallResultToThreadMessage(toolCall model.ToolCall, result string) *thread.Message {
	return thread.NewToolMessage().AddContent(
		thread.NewToolResponseContent(
			thread.ToolResponseData{
				ID:     toolCall.ID,
				Name:   toolCall.Name,
				Result: result,
			},
		),
	)
}

func (o *Cohere) callTools(toolCalls []model.ToolCall) []*thread.Message {
	if len(o.functions) == 0 || len(toolCalls) == 0 {
		return nil
	}

	var messages []*thread.Message
	for _, toolCall := range toolCalls {
		result, err := o.callTool(toolCall)
		if err != nil {
			result = fmt.Sprintf("error: %s", err)
		}

		messages = append(messages, toolCallResultToThreadMessage(toolCall, result))
	}

	return messages
}

func (o *Cohere) callTool(toolCall model.ToolCall) (string, error) {
	fn, ok := o.functions[toolCall.Name]
	if !ok {
		return "", fmt.Errorf("unknown function %s", toolCall.Name)
	}

	parameters, err := json.Marshal(toolCall.Parameters)
	if err != nil {
		return "", err
	}

	resultAsJSON, err := callFnWithArgumentAsJSON(fn.Fn, string(parameters))
	//resultAsJSON, err := callFnWithArgumentAsJSON(fn.Fn, "{}")
	if err != nil {
		return "", err
	}

	return resultAsJSON, nil
}
