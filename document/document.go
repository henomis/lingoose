package document

type Document struct {
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}
