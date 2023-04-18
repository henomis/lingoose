package chat

import (
	"reflect"
	"testing"

	"github.com/henomis/lingoose/prompt/template"
)

func TestChatPromptTemplate_ToMessages(t *testing.T) {
	type fields struct {
		messagesPromptTemplate []MessageTemplate
	}
	type args struct {
		inputs template.Inputs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Message
		wantErr bool
	}{

		{
			name: "TestChatPromptTemplate_ToMessages",
			fields: fields{
				messagesPromptTemplate: []MessageTemplate{
					{
						Type: MessageTypeSystem,
						Template: template.New(
							[]string{"input_language", "output_language"},
							[]string{},
							"You are a helpful assistant that translates {{.input_language}} to {{.output_language}}.",
							nil,
						),
					},
					{
						Type: MessageTypeUser,
						Template: template.New(
							[]string{"text"},
							[]string{},
							"{{.text}}",
							nil,
						),
					},
				},
			},
			args: args{
				inputs: template.Inputs{
					"input_language":  "English",
					"output_language": "French",
					"text":            "I love programming.",
				},
			},
			want: []Message{
				{
					Type:    MessageTypeSystem,
					Content: "You are a helpful assistant that translates English to French.",
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
				messagesPromptTemplate: tt.fields.messagesPromptTemplate,
			}
			got, err := p.ToMessages(tt.args.inputs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChatPromptTemplate.ToMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChatPromptTemplate.ToMessages() = %v, want %v", got, tt.want)
			}
		})
	}
}
