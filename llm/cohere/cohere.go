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

type Model model.Model

const (
	ModelCommand             Model = Model(model.ModelCommand)
	ModelCommandNightly      Model = Model(model.ModelCommandNightly)
	ModelCommandLight        Model = Model(model.ModelCommandLight)
	ModelCommandLightNightly Model = Model(model.ModelCommandLightNightly)
)

const (
	DefaultMaxTokens   = 256
	DefaultTemperature = 0.75
	DefaultModel       = ModelCommand
)

type Cohere struct {
	client      *coherego.Client
	model       Model
	temperature float64
	maxTokens   int
	verbose     bool
	stop        []string
	cache       *cache.Cache
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

// Completion returns the completion for the given prompt
func (c *Cohere) Completion(ctx context.Context, prompt string) (string, error) {
	resp := &response.Generate{}
	err := c.client.Generate(
		ctx,
		&request.Generate{
			Prompt:        prompt,
			Temperature:   &c.temperature,
			MaxTokens:     &c.maxTokens,
			Model:         (*model.Model)(&c.model),
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

func getCacheableMessages(t *thread.Thread) []string {
	userMessages := make([]*thread.Message, 0)
	for _, message := range t.Messages {
		if message.Role == thread.RoleUser {
			userMessages = append(userMessages, message)
		} else {
			userMessages = make([]*thread.Message, 0)
		}
	}

	var messages []string
	for _, message := range userMessages {
		for _, content := range message.Contents {
			if content.Type == thread.ContentTypeText {
				messages = append(messages, content.Data.(string))
			} else {
				messages = make([]string, 0)
				break
			}
		}
	}

	return messages
}

func (o *Cohere) getCache(ctx context.Context, t *thread.Thread) (*cache.Result, error) {
	messages := getCacheableMessages(t)
	cacheQuery := strings.Join(messages, "\n")
	cacheResult, err := o.cache.Get(ctx, cacheQuery)
	if err != nil {
		return cacheResult, err
	}

	t.AddMessage(thread.NewAssistantMessage().AddContent(
		thread.NewTextContent(strings.Join(cacheResult.Answer, "\n")),
	))

	return cacheResult, nil
}

func (o *Cohere) setCache(ctx context.Context, t *thread.Thread, cacheResult *cache.Result) error {
	lastMessage := t.Messages[len(t.Messages)-1]

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

	err := o.cache.Set(ctx, cacheResult.Embedding, strings.Join(contents, "\n"))
	if err != nil {
		return err
	}

	return nil
}

func (o *Cohere) Generate(ctx context.Context, t *thread.Thread) error {
	if t == nil {
		return nil
	}

	var err error
	var cacheResult *cache.Result
	if o.cache != nil {
		cacheResult, err = o.getCache(ctx, t)
		if err == nil {
			return nil
		} else if !errors.Is(err, cache.ErrCacheMiss) {
			return err
		}
	}

	completionQuery := ""
	for _, message := range t.Messages {
		for _, content := range message.Contents {
			if content.Type == thread.ContentTypeText {
				completionQuery += content.Data.(string) + "\n"
			}
		}
	}

	completionResponse, err := o.Completion(ctx, completionQuery)
	if err != nil {
		return err
	}

	t.AddMessage(thread.NewAssistantMessage().AddContent(
		thread.NewTextContent(completionResponse),
	))

	if o.cache != nil {
		err = o.setCache(ctx, t, cacheResult)
		if err != nil {
			return err
		}
	}

	return nil
}
