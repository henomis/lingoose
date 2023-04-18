package template

import (
	"reflect"
	"testing"

	"github.com/henomis/lingoose/prompt/example"
)

func TestNewWithExamples(t *testing.T) {
	type args struct {
		inputs           []string
		outputs          []string
		examples         example.Examples
		examplesTemplate *Template
	}
	tests := []struct {
		name    string
		args    args
		want    *Template
		wantErr bool
	}{
		{
			name: "TestNewWithExamples",
			args: args{
				inputs:  []string{"input"},
				outputs: []string{},
				examples: example.Examples{
					Examples: []example.Example{
						{"word": "happy", "antonym": "sad"},
						{"word": "tall", "antonym": "short"},
					},
					Separator: "\n\n",
					Prefix:    "Give the antonym of every input",
					Suffix:    "Word: {input}\nAntonym:",
				},
				examplesTemplate: &Template{
					inputs:         []string{"word", "antonym"},
					outputs:        []string{},
					template:       "Word: {{.word}}\nAntonym: {{.antonym}}",
					inputsSet:      map[string]struct{}{"word": {}, "antonym": {}},
					templateEngine: nil,
				},
			},
			want: &Template{
				inputs:         []string{"input"},
				outputs:        []string{},
				template:       "Give the antonym of every input\n\nWord: happy\nAntonym: sad\n\nWord: tall\nAntonym: short\n\nWord: {input}\nAntonym:",
				inputsSet:      map[string]struct{}{"input": {}},
				templateEngine: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWithExamples(tt.args.inputs, tt.args.outputs, tt.args.examples, tt.args.examplesTemplate)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWithExamples() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWithExamples() = %v, want %v", got, tt.want)
			}
		})
	}
}
