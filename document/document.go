package document

import "github.com/henomis/lingoose/types"

type Document struct {
	Content  string     `json:"content"`
	Metadata types.Meta `json:"metadata"`
}

// SetMetadata sets the document metadata key to value
func (d *Document) SetMetadata(key string, value interface{}) {
	if d.Metadata == nil {
		d.Metadata = make(types.Meta)
	}
	d.Metadata[key] = value
}

// GetMetadata returns the document metadata
func (d *Document) GetMetadata(key string) (interface{}, bool) {
	value, ok := d.Metadata[key]
	return value, ok
}

// GetContent returns the document content
func (d *Document) GetContent() string {
	return d.Content
}

// GetEnrichedContent returns the document content with the metadata appended
func (d *Document) GetEnrichedContent() string {
	if d.Metadata == nil {
		return d.Content
	}

	return d.Content + "\n\n" + d.Metadata.String()
}
