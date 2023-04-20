package memory

import (
	"reflect"
	"testing"
)

func TestSimpleMemory_Get(t *testing.T) {
	type fields struct {
		memory map[string]interface{}
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name: "Test 1",
			fields: fields{
				memory: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			args: args{
				key: "key1",
			},
			want: "value1",
		},
		{
			name: "Test 2",
			fields: fields{
				memory: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			args: args{
				key: "key3",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SimpleMemory{
				memory: tt.fields.memory,
			}
			if got := m.Get(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SimpleMemory.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleMemory_Set(t *testing.T) {
	type fields struct {
		memory map[string]interface{}
	}
	type args struct {
		key   string
		value interface{}
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
				memory: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			args: args{
				key:   "key3",
				value: "value3",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SimpleMemory{
				memory: tt.fields.memory,
			}
			if err := m.Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("SimpleMemory.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSimpleMemory_All(t *testing.T) {
	type fields struct {
		memory map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]interface{}
	}{
		{
			name: "Test 1",
			fields: fields{
				memory: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SimpleMemory{
				memory: tt.fields.memory,
			}
			if got := m.All(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SimpleMemory.All() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleMemory_Delete(t *testing.T) {
	type fields struct {
		memory map[string]interface{}
	}
	type args struct {
		key string
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
				memory: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			args: args{
				key: "key1",
			},
			wantErr: false,
		},
		{
			name: "Test 1",
			fields: fields{
				memory: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			args: args{
				key: "key3",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SimpleMemory{
				memory: tt.fields.memory,
			}
			if err := m.Delete(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("SimpleMemory.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSimpleMemory_Clear(t *testing.T) {
	type fields struct {
		memory map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Test 1",
			fields: fields{
				memory: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SimpleMemory{
				memory: tt.fields.memory,
			}
			if err := m.Clear(); (err != nil) != tt.wantErr {
				t.Errorf("SimpleMemory.Clear() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
