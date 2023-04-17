package prompt

import (
	"reflect"
	"testing"
	"text/template"
)

func TestPromptTemplate_Format(t *testing.T) {
	type fields struct {
		Inputs    []string
		Outputs   []string
		Template  string
		inputsSet map[string]struct{}
		template  *template.Template
	}
	type args struct {
		promptTemplateInputs PromptTemplateInputs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "TestPromptTemplate_Format empty",
			fields: fields{
				Inputs:    []string{},
				Outputs:   []string{},
				Template:  "Tell me a joke.",
				inputsSet: map[string]struct{}{},
				template:  nil,
			},
			args: args{
				promptTemplateInputs: PromptTemplateInputs{},
			},
			want:    "Tell me a joke.",
			wantErr: false,
		},
		{
			name: "TestPromptTemplate_Format one input",
			fields: fields{
				Inputs:    []string{"name"},
				Outputs:   []string{},
				Template:  "Tell me a joke about {{.name}}.",
				inputsSet: map[string]struct{}{"name": {}},
				template:  nil,
			},
			args: args{
				promptTemplateInputs: PromptTemplateInputs{
					"name": "llamas",
				},
			},
			want:    "Tell me a joke about llamas.",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PromptTemplate{
				Inputs:    tt.fields.Inputs,
				Outputs:   tt.fields.Outputs,
				Template:  tt.fields.Template,
				inputsSet: tt.fields.inputsSet,
				template:  tt.fields.template,
			}
			got, err := p.Format(tt.args.promptTemplateInputs)
			if (err != nil) != tt.wantErr {
				t.Errorf("PromptTemplate.Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PromptTemplate.Format() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPromptTemplate_NewFromLangchain(t *testing.T) {
	type fields struct {
		Inputs    []string
		Outputs   []string
		Template  string
		inputsSet map[string]struct{}
		template  *template.Template
	}
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *PromptTemplate
		wantErr bool
	}{
		{
			name: "TestPromptTemplate_NewFromLangchain",
			fields: fields{
				Inputs:    []string{},
				Outputs:   []string{},
				Template:  "",
				inputsSet: map[string]struct{}{},
				template:  nil,
			},
			args: args{
				url: "lc://prompts/conversation/prompt.json",
			},
			want: &PromptTemplate{
				Inputs:    []string{"history", "input"},
				Outputs:   []string{},
				Template:  "The following is a friendly conversation between a human and an AI. The AI is talkative and provides lots of specific details from its context. If the AI does not know the answer to a question, it truthfully says it does not know.\n\nCurrent conversation:\n{{.history}}\nHuman: {{.input}}\nAI:",
				inputsSet: map[string]struct{}{"history": {}, "input": {}},
				template:  nil,
			},
			wantErr: false,
		},
		{
			name: "TestPromptTemplate_NewFromLangchain",
			fields: fields{
				Inputs:    []string{},
				Outputs:   []string{},
				Template:  "",
				inputsSet: map[string]struct{}{},
				template:  nil,
			},
			args: args{
				url: "lc://prompts/hello-world/prompt.yaml",
			},
			want: &PromptTemplate{
				Inputs:    []string{},
				Outputs:   []string{},
				Template:  "Say hello world.",
				inputsSet: map[string]struct{}{},
				template:  nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PromptTemplate{
				Inputs:    tt.fields.Inputs,
				Outputs:   tt.fields.Outputs,
				Template:  tt.fields.Template,
				inputsSet: tt.fields.inputsSet,
				template:  tt.fields.template,
			}
			got, err := p.NewFromLangchain(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("PromptTemplate.NewFromLangchain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PromptTemplate.NewFromLangchain() = %v, want %v", got, tt.want)
			}
		})
	}
}
