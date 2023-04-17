package prompt

import (
	"reflect"
	"testing"
)

func TestChatPromptTemplate_ToMessages(t *testing.T) {
	type fields struct {
		messagesPromptTemplate []MessagePromptTemplate
	}
	type args struct {
		inputs Inputs
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
				messagesPromptTemplate: []MessagePromptTemplate{
					{
						Type: MessageTypeSystem,
						Prompt: &PromptTemplate{
							template: "You are a helpful assistant that translates {{.input_language}} to {{.output_language}}.",
							inputs:   []string{"input_language", "output_language"},
							inputsSet: map[string]struct{}{
								"input_language":  {},
								"output_language": {},
							},
						},
					},
					{
						Type: MessageTypeUser,
						Prompt: &PromptTemplate{
							template: "{{.text}}",
							inputs:   []string{"text"},
							inputsSet: map[string]struct{}{
								"text": {},
							},
						},
					},
				},
			},
			args: args{
				inputs: Inputs{
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
			p := &ChatPromptTemplate{
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
