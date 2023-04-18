// Package prompt provides a easy way to format a prompt using the Go template engine.
// Prompts are defined using a template string and a list of inputs.
package prompt

import (
	"testing"
	texttemplate "text/template"
)

type simpleStructInput struct {
	Name string `json:"name" validate:"required,max=5"`
}

type simpleStructOutput struct {
	Value string `json:"value"`
}

var helloInput = simpleStructInput{
	Name: "world",
}

var helloInputMax = simpleStructInput{
	Name: "worldworld",
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
		OutputDecoder  OutputDecoderFn
		Template       string
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
				Template:       "Tell me a joke.",
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
				Template:       "Hello {{.Name}}",
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
				Template:       "Hello {{.Name.First}} {{.Name.Middle}} {{.Lastname}}",
				templateEngine: nil,
			},
			want:    "Hello Alan Mathison Turing",
			wantErr: false,
		},
		{
			name: "TestPrompt_Format hello simple",
			fields: fields{
				Input:          helloInputMax,
				Output:         nil,
				OutputDecoder:  nil,
				Template:       "Hello {{.Name}}",
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
				Template:       "Hello {{.Lastname}}",
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
				Template:       "Hello {{}}",
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
