package openai

import (
	"reflect"
	"testing"

	"github.com/henomis/lingoose/thread"
	"github.com/sashabaranov/go-openai"
)

func TestWithResponseFormatJSONSchema(t *testing.T) {
	tests := []struct {
		name       string
		jsonSchema *ResponseFormatJSONSchema
		want       *ResponseFormatJSONSchema
	}{
		{
			name: "should set JSON schema",
			jsonSchema: &ResponseFormatJSONSchema{
				Name:        "FirstName",
				Description: "First of your name",
				Strict:      true,
			},
		},
		{
			name:       "should set nil JSON schema",
			jsonSchema: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := New().
				WithResponseFormatJSONSchema(tt.jsonSchema).
				WithResponseFormat(openai.ChatCompletionResponseFormatTypeJSONSchema)
			if !reflect.DeepEqual(tt.jsonSchema, o.responseFormatJSONSchema) {
				t.Fatalf("New OpenAI WithResponseFormatJSONSchema doesn't set value correctly")
			}

			myThread := thread.New()
			request := o.buildChatCompletionRequest(myThread)
			if !reflect.DeepEqual(tt.jsonSchema, request.ResponseFormat.JSONSchema) {
				t.Fatalf("New OpenAI WithResponseFormatJSONSchema doesn't generate correct chat request correctly")
			}
		})
	}
}
