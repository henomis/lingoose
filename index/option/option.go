package option

type Option func(*Options)

type Options struct {
	TopK   int
	Filter any
}

func WithTopK(topK int) Option {
	return func(opts *Options) {
		opts.TopK = topK
	}
}

func WithFilter(filter any) Option {
	return func(opts *Options) {
		opts.Filter = filter
	}
}
