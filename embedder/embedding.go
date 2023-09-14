package embedder

var (
	ErrCreateEmbedding = "unable to create embedding"
)

// Embedding is the result of an embedding operation.
type Embedding []float64

func (e Embedding) ToFloat32() []float32 {
	vect := make([]float32, len(e))
	for i, v := range e {
		vect[i] = float32(v)
	}

	return vect
}
