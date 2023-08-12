package option

type Option func(*Options)

type Options struct {
	TopK   int
	Filter interface{}
}

func WithTopK(topK int) Option {
	return func(opts *Options) {
		opts.TopK = topK
	}
}

func WithFilter(filter interface{}) Option {
	return func(opts *Options) {
		opts.Filter = filter
	}
}
