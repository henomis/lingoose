package chat

import (
	"reflect"
	"testing"

	"github.com/henomis/lingoose/prompt"
)

func TestChat_ToMessages(t *testing.T) {
	type fields struct {
		PromptMessages PromptMessages
	}
	tests := []struct {
		name    string
		fields  fields
		want    Messages
		wantErr bool
	}{
		{
			name: "TestChat_ToMessages",
			fields: fields{
				PromptMessages: PromptMessages{
					{
						Type: MessageTypeSystem,
						Prompt: &prompt.Prompt{
							Input: map[string]interface{}{
								"input_language":  "English",
								"output_language": "Spanish",
							},
							Template: "You are a helpful assistant that translates {{.input_language}} to {{.output_language}}.",
						},
					},
					{
						Type: MessageTypeUser,
						Prompt: &prompt.Prompt{
							Input: map[string]interface{}{
								"text": "I love programming.",
							},
							Template: "{{.text}}",
						},
					},
				},
			},
			want: Messages{
				{
					Type:    MessageTypeSystem,
					Content: "You are a helpful assistant that translates English to Spanish.",
				},
				{
					Type:    MessageTypeUser,
					Content: "I love programming.",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Chat{
				PromptMessages: tt.fields.PromptMessages,
			}
			got, err := p.ToMessages()
			if (err != nil) != tt.wantErr {
				t.Errorf("Chat.ToMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Chat.ToMessages() = %v, want %v", got, tt.want)
			}
		})
	}
}
