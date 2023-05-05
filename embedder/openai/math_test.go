package openaiembedder

import (
	"reflect"
	"testing"

	"github.com/henomis/lingoose/embedder"
)

func Test_norm(t *testing.T) {
	type args struct {
		a []float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "norm 1",
			args: args{
				a: []float64{1, 2, 3},
			},
			want: 3.7416573867739413,
		},
		{
			name: "norm 2",
			args: args{
				a: []float64{1},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := norm(tt.args.a); got != tt.want {
				t.Errorf("norm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_average(t *testing.T) {
	type args struct {
		embeddings []embedder.Embedding
		lens       []float64
	}
	tests := []struct {
		name string
		args args
		want []float64
	}{
		{
			name: "average 1",
			args: args{
				embeddings: []embedder.Embedding{
					{1, 2, 3},
					{4, 5, 6},
				},
				lens: []float64{1, 1},
			},
			want: []float64{2.5, 3.5, 4.5},
		},
		{
			name: "average 2",
			args: args{
				embeddings: []embedder.Embedding{
					{1, 2, 3},
					{4, 5, 6},
				},
				lens: []float64{2, 1},
			},
			want: []float64{2, 3, 4},
		},
		{
			name: "average 3",
			args: args{
				embeddings: []embedder.Embedding{
					{1, 2, 3},
					{4, 5, 6},
				},
				lens: []float64{1, 2},
			},
			want: []float64{3, 4, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := average(tt.args.embeddings, tt.args.lens); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("average() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_normalizeEmbeddings(t *testing.T) {
	type args struct {
		embeddings []embedder.Embedding
		lens       []float64
	}
	tests := []struct {
		name string
		args args
		want []float64
	}{
		{
			name: "normalize 1",
			args: args{
				embeddings: []embedder.Embedding{
					{1, 2, 3},
					{4, 5, 6},
				},
				lens: []float64{1, 1},
			},
			want: []float64{0.4016096644512494, 0.5622535302317492, 0.722897396012249},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeEmbeddings(tt.args.embeddings, tt.args.lens); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalizeEmbeddings() = %v, want %v", got, tt.want)
			}
		})
	}
}
