package template

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
		promptTemplateInputs Inputs
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
				promptTemplateInputs: Inputs{},
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
				promptTemplateInputs: Inputs{
					"name": "llamas",
				},
			},
			want:    "Tell me a joke about llamas.",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Prompt{
				inputs:         tt.fields.Inputs,
				outputs:        tt.fields.Outputs,
				template:       tt.fields.Template,
				inputsSet:      tt.fields.inputsSet,
				templateEngine: tt.fields.template,
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

func TestNewFromLangchain(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    *Prompt
		wantErr bool
	}{
		{
			name: "TestNewFromLangchain",
			args: args{
				url: "lc://prompts/summarize/stuff/prompt.yaml",
			},
			want: &Prompt{
				inputs:         []string{"text"},
				outputs:        []string{},
				template:       "Write a concise summary of the following:\n\n{{.text}}\n\nCONCISE SUMMARY:",
				inputsSet:      map[string]struct{}{"text": {}},
				templateEngine: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFromLangchain(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromLangchain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFromLangchain() = %v, want %v", got, tt.want)
			}
		})
	}
}
