// Package decoder provides a set of decoders to decode the output of a command
package decoder

import (
	"reflect"
	"testing"

	"github.com/rsest/lingoose/types"
)

func TestJSONDecoder_Decode(t *testing.T) {
	type fields struct {
		output types.M
	}
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    types.M
		wantErr bool
	}{
		{
			name: "TestJSONDecoder_Decode",
			fields: fields{
				output: types.M{},
			},
			args: args{
				input: `{"test": "test"}`,
			},
			want: types.M{
				types.DefaultOutputKey: types.M{
					"test": "test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &JSONDecoder{
				output: tt.fields.output,
			}
			got, err := d.Decode(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONDecoder.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JSONDecoder.Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegExDecoder_Decode(t *testing.T) {
	type fields struct {
		output types.M
		regex  string
	}
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    types.M
		wantErr bool
	}{
		{
			name: "TestRegExDecoder_Decode",
			fields: fields{
				output: types.M{},
				regex:  `([a-z]+)(\d+)`,
			},
			args: args{
				input: `test123`,
			},
			want: types.M{
				types.DefaultOutputKey: []string{"test", "123"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &RegExDecoder{
				output: tt.fields.output,
				regex:  tt.fields.regex,
			}
			got, err := d.Decode(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegExDecoder.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RegExDecoder.Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}
