package index

type Option func(*options)

type options struct {
	topK int
}

func WithTopK(topK int) Option {
	return func(opts *options) {
		opts.topK = topK
	}
}
