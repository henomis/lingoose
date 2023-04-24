package embedding

type EmbeddingObject struct {
	Vector []float32 `json:"vector"`
	Index  int       `json:"index"`
}
