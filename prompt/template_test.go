package prompt

import (
	"testing"
	texttemplate "text/template"
)

func TestPromptTemplate_Format(t *testing.T) {
	type fields struct {
		input          interface{}
		template       string
		value          string
		templateEngine *texttemplate.Template
	}
	type args struct {
		input interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test 1",
			fields: fields{
				input:    map[string]interface{}{},
				template: "Hello {{.name}}",
			},
			args: args{
				input: map[string]interface{}{
					"name": "John",
				},
			},
			wantErr: false,
		},
		{
			name: "Test 2",
			fields: fields{
				input:    map[string]interface{}{"name": "Alan"},
				template: "Hello {{.name}}, i'm {{.age}} years old",
			},
			args: args{
				input: map[string]interface{}{
					"age": 30,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PromptTemplate{
				input:          tt.fields.input,
				template:       tt.fields.template,
				value:          tt.fields.value,
				templateEngine: tt.fields.templateEngine,
			}
			if err := p.Format(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("PromptTemplate.Format() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPromptTemplate_Prompt(t *testing.T) {
	type fields struct {
		input          interface{}
		template       string
		value          string
		templateEngine *texttemplate.Template
		external       interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Test 1",
			fields: fields{
				input:    map[string]interface{}{},
				template: "Hello {{.name}}",
				external: map[string]interface{}{
					"name": "John",
				},
			},
			want: "Hello John",
		},
		{
			name: "Test 2",
			fields: fields{
				input:    map[string]interface{}{"age": "33"},
				template: "Hello {{.name}}, i'm {{.age}} years old",
				external: map[string]interface{}{
					"name": "John",
				},
			},
			want: "Hello John, i'm 33 years old",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PromptTemplate{
				input:          tt.fields.input,
				template:       tt.fields.template,
				value:          tt.fields.value,
				templateEngine: tt.fields.templateEngine,
			}

			if tt.fields.external != nil {
				err := p.Format(tt.fields.external)
				if err != nil {
					t.Errorf("PromptTemplate.Prompt() error = %v", err)
				}
			}

			if got := p.Prompt(); got != tt.want {
				t.Errorf("PromptTemplate.Prompt() = %v, want %v", got, tt.want)
			}
		})
	}
}
