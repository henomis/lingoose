package thread

import (
	"reflect"
	"testing"

	"github.com/henomis/lingoose/types"
)

func TestContent_Format(t *testing.T) {
	type fields struct {
		Type ContentType
		Data any
	}
	type args struct {
		input types.M
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Content
	}{
		{
			name: "Test 1",
			fields: fields{
				Type: ContentTypeText,
				Data: "Hello, World!",
			},
			args: args{
				input: types.M{
					"key": "value",
				},
			},
			want: &Content{
				Type: ContentTypeText,
				Data: "Hello, World!",
			},
		},
		{
			name: "Test 2",
			fields: fields{
				Type: ContentTypeText,
				Data: "Hello, {{.key}}!",
			},
			args: args{
				input: types.M{
					"key": "World",
				},
			},
			want: &Content{
				Type: ContentTypeText,
				Data: "Hello, World!",
			},
		},
		{
			name: "Test 3",
			fields: fields{
				Type: ContentTypeText,
				Data: "{{.text}}, {{.key}}!",
			},
			args: args{
				input: types.M{
					"key":  "World",
					"text": "Hello",
				},
			},
			want: &Content{
				Type: ContentTypeText,
				Data: "Hello, World!",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Content{
				Type: tt.fields.Type,
				Data: tt.fields.Data,
			}
			if got := c.Format(tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Content.Format() = %v, want %v", got, tt.want)
			}
		})
	}
}
