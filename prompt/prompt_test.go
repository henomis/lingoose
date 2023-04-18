// Package prompt provides a easy way to format a prompt using the Go template engine.
// Prompts are defined using a template string and a list of inputs.
package prompt

import (
	"testing"
	texttemplate "text/template"

	"github.com/henomis/lingoose/prompt/decoder"
)

type simpleStructInput struct {
	Name string `json:"name"`
}

type simpleStructOutput struct {
	Value string `json:"value"`
}

var helloInput = simpleStructInput{
	Name: "world",
}

var helloOutput simpleStructOutput

type complexStructInput struct {
	Name struct {
		First  string `json:"first"`
		Middle string `json:"middle"`
	} `json:"name"`
	Lastname string `json:"lastname"`
}

var complexInput = complexStructInput{
	Name: struct {
		First  string `json:"first"`
		Middle string `json:"middle"`
	}{
		First:  "Alan",
		Middle: "Mathison",
	},
	Lastname: "Turing",
}

func TestPrompt_Format(t *testing.T) {
	type fields struct {
		Input          interface{}
		Output         interface{}
		OutputDecoder  decoder.DecoderFn
		Template       *string
		templateEngine *texttemplate.Template
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "TestPrompt_Format empty",
			fields: fields{
				Input:          nil,
				Output:         nil,
				OutputDecoder:  nil,
				Template:       newString("Tell me a joke."),
				templateEngine: nil,
			},
			want:    "Tell me a joke.",
			wantErr: false,
		},
		{
			name: "TestPrompt_Format hello simple",
			fields: fields{
				Input:          helloInput,
				Output:         nil,
				OutputDecoder:  nil,
				Template:       newString("Hello {{.Name}}"),
				templateEngine: nil,
			},
			want:    "Hello world",
			wantErr: false,
		},
		{
			name: "TestPrompt_Format hello complex",
			fields: fields{
				Input:          complexInput,
				Output:         nil,
				OutputDecoder:  nil,
				Template:       newString("Hello {{.Name.First}} {{.Name.Middle}} {{.Lastname}}"),
				templateEngine: nil,
			},
			want:    "Hello Alan Mathison Turing",
			wantErr: false,
		},
		{
			name: "TestPrompt_Format hello simple",
			fields: fields{
				Input:          helloInput,
				Output:         nil,
				OutputDecoder:  nil,
				Template:       newString("Hello {{.Lastname}}"),
				templateEngine: nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "TestPrompt_Format hello simple",
			fields: fields{
				Input:          helloInput,
				Output:         nil,
				OutputDecoder:  nil,
				Template:       newString("Hello {{}}"),
				templateEngine: nil,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Prompt{
				Input:          tt.fields.Input,
				Output:         tt.fields.Output,
				OutputDecoder:  tt.fields.OutputDecoder,
				Template:       tt.fields.Template,
				templateEngine: tt.fields.templateEngine,
			}
			got, err := p.Format()
			if (err != nil) != tt.wantErr {
				t.Errorf("Prompt.Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Prompt.Format() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newString(s string) *string {
	return &s
}
