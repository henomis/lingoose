package pipeline

import (
	"reflect"
	"testing"

	"github.com/henomis/lingoose/types"
)

func Test_mergeMaps(t *testing.T) {
	type args struct {
		m1 types.M
		m2 types.M
	}
	tests := []struct {
		name string
		args args
		want types.M
	}{
		{
			name: "Test 1",
			args: args{
				m1: types.M{
					"key1": "value1",
					"key2": "value2",
				},
				m2: types.M{
					"key3": "value3",
					"key4": "value4",
				},
			},
			want: types.M{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeMaps(tt.args.m1, tt.args.m2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeMaps() = %v, want %v", got, tt.want)
			}
		})
	}
}
