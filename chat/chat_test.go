package chat

import (
	"reflect"
	"testing"

	"github.com/henomis/lingoose/prompt"
	"github.com/henomis/lingoose/types"
)

func TestChat_ToMessages(t *testing.T) {

	prompt1, _ := prompt.NewPromptTemplate(
		"You are a helpful assistant that translates {{.input_language}} to {{.output_language}}.",
		types.M{
			"input_language":  "English",
			"output_language": "Spanish",
		},
	)

	prompt2 := prompt.New("What is your name?")

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
						Type:   MessageTypeSystem,
						Prompt: prompt1,
					},
					{
						Type:   MessageTypeUser,
						Prompt: prompt2,
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
					Content: "What is your name?",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Chat{
				promptMessages: tt.fields.PromptMessages,
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
