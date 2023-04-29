package embedder

var (
	ErrCreateEmbedding = "unable to create embedding"
)

// Embedding is the result of an embedding operation.
type Embedding struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}
