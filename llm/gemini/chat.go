package gemini

import (
	"cloud.google.com/go/vertexai/genai"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/henomis/lingoose/thread"
	"google.golang.org/api/iterator"
	"strings"
)

func (g *Gemini) streamChat(ctx context.Context, t *thread.Thread, parts []genai.Part) error {

	iter := g.session.SendMessageStream(ctx, parts...)

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
		if response.Candidates[0].Content != nil {
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
	}
	t.AddMessages(messages...)
	return nil
}

func (g *Gemini) generateChat(ctx context.Context, t *thread.Thread, parts []genai.Part) error {

	response, err := g.session.SendMessage(ctx, parts...)
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
