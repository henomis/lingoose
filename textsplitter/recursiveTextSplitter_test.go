package textsplitter

import (
	"reflect"
	"testing"
	"unicode/utf8"

	"github.com/rsest/lingoose/document"
	"github.com/rsest/lingoose/types"
)

//nolint:dupword,funlen
func TestRecursiveCharacterTextSplitter_SplitDocuments(t *testing.T) {
	type fields struct {
		textSplitter TextSplitter
		separators   []string
	}
	type args struct {
		documents []document.Document
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []document.Document
	}{
		{
			name: "TestRecursiveCharacterTextSplitter_SplitDocuments",
			fields: fields{
				textSplitter: TextSplitter{
					chunkSize:    10,
					chunkOverlap: 0,
					lengthFunction: func(s string) int {
						return len(s)
					},
				},
				separators: []string{"\n\n", "\n", " ", ""},
			},
			args: args{
				documents: []document.Document{
					{
						Content:  "This is a test",
						Metadata: types.Meta{},
					},
				},
			},
			want: []document.Document{
				{
					Content:  "This is a",
					Metadata: types.Meta{},
				},
				{
					Content:  "test",
					Metadata: types.Meta{},
				},
			},
		},
		{
			name: "TestRecursiveCharacterTextSplitter_SplitDocuments",
			fields: fields{
				textSplitter: TextSplitter{
					chunkSize:    20,
					chunkOverlap: 1,
					lengthFunction: func(s string) int {
						return len(s)
					},
				},
				separators: []string{"\n", "$"},
			},
			args: args{
				documents: []document.Document{
					{
						Content:  "Hi, Harrison. \nI am glad to meet you",
						Metadata: types.Meta{},
					},
				},
			},
			want: []document.Document{
				{
					Content:  "Hi, Harrison.",
					Metadata: types.Meta{},
				},
				{
					Content:  "I am glad to meet you",
					Metadata: types.Meta{},
				},
			},
		},
		{
			name: "TestRecursiveCharacterTextSplitter_SplitDocuments",
			fields: fields{
				textSplitter: TextSplitter{
					chunkSize:      10,
					chunkOverlap:   0,
					lengthFunction: utf8.RuneCountInString,
				},
				separators: []string{"\n\n", "\n", " "},
			},
			args: args{
				documents: []document.Document{
					{
						Content:  "哈里森\n很高兴遇见你\n欢迎来中国",
						Metadata: types.Meta{},
					},
				},
			},
			want: []document.Document{
				{
					Content:  "哈里森\n很高兴遇见你",
					Metadata: types.Meta{},
				},
				{
					Content:  "欢迎来中国",
					Metadata: types.Meta{},
				},
			},
		},
		{
			name: "TestRecursiveCharacterTextSplitter_SplitDocuments",
			fields: fields{
				textSplitter: TextSplitter{
					chunkSize:    10,
					chunkOverlap: 0,
					lengthFunction: func(s string) int {
						return len(s)
					},
				},
				separators: []string{"\n\n", "\n", " ", ""},
			},
			args: args{
				documents: []document.Document{
					{
						Content: "This is a test",
						Metadata: types.Meta{
							"test":  "test",
							"test2": "test2",
						},
					},
				},
			},
			want: []document.Document{
				{
					Content: "This is a",
					Metadata: types.Meta{
						"test":  "test",
						"test2": "test2",
					},
				},
				{
					Content: "test",
					Metadata: types.Meta{
						"test":  "test",
						"test2": "test2",
					},
				},
			},
		},
		{
			name: "TestRecursiveCharacterTextSplitter_SplitDocuments2",
			fields: fields{
				textSplitter: TextSplitter{
					chunkSize:    20,
					chunkOverlap: 5,
					lengthFunction: func(s string) int {
						return len(s)
					},
				},
				separators: []string{"\n\n", "\n", " ", ""},
			},
			args: args{
				documents: []document.Document{
					{
						Content:  "Lorem ipsum dolor sit amet,\n\nconsectetur adipisci elit",
						Metadata: types.Meta{},
					},
				},
			},
			want: []document.Document{
				{
					Content:  "Lorem ipsum dolor",
					Metadata: types.Meta{},
				},
				{
					Content:  "dolor sit amet,",
					Metadata: types.Meta{},
				},
				{
					Content:  "consectetur adipisci",
					Metadata: types.Meta{},
				},
				{
					Content:  "elit",
					Metadata: types.Meta{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RecursiveCharacterTextSplitter{
				TextSplitter: tt.fields.textSplitter,
				separators:   tt.fields.separators,
			}
			if got := r.SplitDocuments(tt.args.documents); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RecursiveCharacterTextSplitter.SplitDocuments() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
