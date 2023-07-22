package index

type Option func(*options)

type options struct {
	topK   int
	filter interface{}
}

func WithTopK(topK int) Option {
	return func(opts *options) {
		opts.topK = topK
	}
}

func WithFilter(filter interface{}) Option {
	return func(opts *options) {
		opts.filter = filter
	}
}
