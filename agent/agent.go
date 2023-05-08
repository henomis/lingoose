package agent

import (
	"context"
	"fmt"
	"regexp"

	"github.com/henomis/lingoose/chat"
	"github.com/henomis/lingoose/pipeline"
	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
)

const (
	actionRegexExpr      = `Action: (.*)`
	actionInputRegexExpr = `Action Input: "?(.*)"?`
	finalAnswerRegexExpr = `Final Answer: (.*)`
)

type Llm interface {
	Completion(ctx context.Context, prompt string) (string, error)
	Chat(ctx context.Context, chat *chat.Chat) (string, error)
}

type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, input string) (string, error)
}

type agent struct {
	name              string
	llm               Llm
	tube              *pipeline.Tube
	tools             map[string]Tool
	actionRegexp      *regexp.Regexp
	actionInputRegexp *regexp.Regexp
	finalAnswerRegexp *regexp.Regexp
}

func New(name string, llm Llm, tools []Tool) (*agent, error) {

	toolMap := make(map[string]Tool)
	for _, tool := range tools {
		toolMap[tool.Name()] = tool
	}

	prompt, err := prompt.NewPromptTemplate(
		promptTemplate,
		types.M{
			"tools": tools,
		},
	)
	if err != nil {
		return nil, err
	}

	tubeLlm := pipeline.Llm{
		LlmEngine: llm,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt,
		Chat:      nil,
	}

	tube := pipeline.NewTube(
		"agent-"+name,
		tubeLlm,
		nil,
		nil,
	)

	actionRegexp, err := regexp.Compile(actionRegexExpr)
	if err != nil {
		return nil, err
	}

	actionInputRegexExpr, err := regexp.Compile(actionInputRegexExpr)
	if err != nil {
		return nil, err
	}

	finalAnswerRegexp, err := regexp.Compile(finalAnswerRegexExpr)
	if err != nil {
		return nil, err
	}

	return &agent{
		name:              "agent-" + name,
		llm:               llm,
		tube:              tube,
		tools:             toolMap,
		actionRegexp:      actionRegexp,
		actionInputRegexp: actionInputRegexExpr,
		finalAnswerRegexp: finalAnswerRegexp,
	}, nil

}

func (a *agent) Run(ctx context.Context, input types.M) (types.M, error) {

	prompt := ""

	for {
		res, err := a.tube.Run(ctx, input)
		if err != nil {
			return nil, err
		}

		output := res["output"].(string)

		// add this to the prompt
		prompt += output

		action := a.actionRegexp.FindStringSubmatch(output)
		if len(action) == 2 {
			toolName := action[1]
			actionInput := a.actionInputRegexp.FindStringSubmatch(output)
			if len(actionInput) == 2 {

				tool, ok := a.tools[toolName]
				if !ok {
					return nil, fmt.Errorf("tool %s not found", toolName)
				}

				toolOutput, err := tool.Execute(ctx, actionInput[1])
				if err != nil {
					return nil, err
				}

				prompt += fmt.Sprintf("\nObservation: %s\n", toolOutput)

			}

		}

		finalAnswer := a.finalAnswerRegexp.FindStringSubmatch(output)
		if len(finalAnswer) == 2 {
			res["output"] = finalAnswer[1]
			return res, nil
		}

		a.rebuildPrompt(prompt)

	}

}

func (a *agent) rebuildPrompt(text string) error {
	prompt, err := prompt.NewPromptTemplate(
		promptTemplate+text,
		types.M{
			"tools": a.tools,
		},
	)
	if err != nil {
		return err
	}

	tubeLlm := pipeline.Llm{
		LlmEngine: a.llm,
		LlmMode:   pipeline.LlmModeCompletion,
		Prompt:    prompt,
		Chat:      nil,
	}

	a.tube = pipeline.NewTube(
		a.name,
		tubeLlm,
		nil,
		nil,
	)

	return nil
}
