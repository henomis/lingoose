package openaiembedder

import (
	"reflect"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func Test_openAIEmbedder_splitText(t *testing.T) {
	type fields struct {
		openAIClient *openai.Client
		model        Model
	}
	type args struct {
		text      string
		maxTokens int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "split text 1",
			fields: fields{
				openAIClient: nil,
				model:        AdaEmbeddingV2,
			},
			args: args{
				text:      "Hello, my name is John",
				maxTokens: 5,
			},
			want:    []string{"Hello, my name is", " John"},
			wantErr: false,
		},
		{
			name: "split text 2",
			fields: fields{
				openAIClient: nil,
				model:        AdaEmbeddingV2,
			},
			args: args{
				text:      "Hello, my name is John",
				maxTokens: 100,
			},
			want:    []string{"Hello, my name is John"},
			wantErr: false,
		},
		{
			name: "split text 2",
			fields: fields{
				openAIClient: nil,
				model:        AdaEmbeddingV2,
			},
			args: args{
				text: "Lorem ipsum dolor sit amet, consectetur adipisci elit, sed do eiusmod tempor incidunt" +
					" ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrum exercitationem ullamco " +
					"laboriosam, nisi ut aliquid ex ea commodi consequatur. Duis aute irure reprehenderit in voluptate " +
					"velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint obcaecat cupiditat non proident, " +
					"sunt in culpa qui officia deserunt mollit anim id est laborum.",
				maxTokens: 10,
			},
			want: []string{
				"Lorem ipsum dolor sit amet, consectetur adipisci elit",
				", sed do eiusmod tempor incidunt ut labore et",
				" dolore magna aliqua. Ut enim ad minim veniam,",
				" quis nostrum exercitationem ullamco laboriosam",
				", nisi ut aliquid ex ea commodi consequ",
				"atur. Duis aute irure reprehenderit in volupt",
				"ate velit esse cillum dolore eu fugiat nulla",
				" pariatur. Excepteur sint obcaecat",
				" cupiditat non proident, sunt in culpa qui",
				" officia deserunt mollit anim id est labor",
				"um.",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OpenAIEmbedder{
				openAIClient: tt.fields.openAIClient,
				model:        tt.fields.model,
			}
			got, err := o.chunkText(tt.args.text, tt.args.maxTokens)
			if (err != nil) != tt.wantErr {
				t.Errorf("openAIEmbedder.splitText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("openAIEmbedder.splitText() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
