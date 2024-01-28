package cohere

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	coherego "github.com/henomis/cohere-go"
	"github.com/henomis/cohere-go/model"
	"github.com/henomis/cohere-go/request"
	"github.com/henomis/cohere-go/response"
	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/llm/cache"
	"github.com/henomis/lingoose/thread"
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

	if c.streamCallbackFn != nil {
		err = c.stream(ctx, t, chatRequest)
	} else {
		err = c.generate(ctx, t, chatRequest)
	}

	if err != nil {
		return err
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

	t.AddMessage(thread.NewAssistantMessage().AddContent(
		thread.NewTextContent(response.Text),
	))

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
