package prompt

import (
	"testing"

	"github.com/rsest/lingoose/types"
)

func TestPromptTemplate_Format(t *testing.T) {
	type fields struct {
		input    interface{}
		template string
	}
	type args struct {
		input types.M
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
				input:    types.M{},
				template: "Hello {{.name}}",
			},
			args: args{
				input: types.M{
					"name": "John",
				},
			},
			wantErr: false,
		},
		{
			name: "Test 2",
			fields: fields{
				input:    types.M{"name": "Alan"},
				template: "Hello {{.name}}, i'm {{.age}} years old",
			},
			args: args{
				input: types.M{
					"age": 30,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPromptTemplate(tt.fields.template).WithInputs(tt.fields.input)

			if err := p.Format(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("PromptTemplate.Format() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPromptTemplate_Prompt(t *testing.T) {
	type fields struct {
		input    interface{}
		template string
		external types.M
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Test 1",
			fields: fields{
				input:    types.M{},
				template: "Hello {{.name}}",
				external: types.M{
					"name": "John",
				},
			},
			want: "Hello John",
		},
		{
			name: "Test 2",
			fields: fields{
				input:    types.M{"age": "33"},
				template: "Hello {{.name}}, i'm {{.age}} years old",
				external: types.M{
					"name": "John",
				},
			},
			want: "Hello John, i'm 33 years old",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPromptTemplate(tt.fields.template).WithInputs(tt.fields.input)

			if tt.fields.external != nil {
				err := p.Format(tt.fields.external)
				if err != nil {
					t.Errorf("PromptTemplate.Prompt() error = %v", err)
				}
			}

			if got := p.String(); got != tt.want {
				t.Errorf("PromptTemplate.Prompt() = %v, want %v", got, tt.want)
			}
		})
	}
}
