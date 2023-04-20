package pipeline

import (
	"reflect"
	"testing"
)

func Test_mergeMaps(t *testing.T) {
	type args struct {
		m1 map[string]interface{}
		m2 map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "Test 1",
			args: args{
				m1: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
				m2: map[string]interface{}{
					"key3": "value3",
					"key4": "value4",
				},
			},
			want: map[string]interface{}{
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
