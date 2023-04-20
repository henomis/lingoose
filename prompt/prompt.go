package prompt

import "errors"

var (
	ErrFormatting     = errors.New("formatting prompt error")
	ErrDecoding       = errors.New("decoding input error")
	ErrTemplateEngine = errors.New("template engine error")
)
