package websocket

type Option interface {
	apply(bk *breaker)
}

type OptionFunc func(bk *breaker)

func (f OptionFunc) apply(bk *breaker) { f(bk) }

// WithErrorHandlerOption returns a function which sets error handler.
func WithErrorHandlerOption(f ErrorHandler) OptionFunc {
	return func(bk *breaker) {
		bk.errorHandler = f
	}
}

// WithMaxReadLimit returns a function which sets max size how many data reads from websocket.
func WithMaxReadLimit(length int64) OptionFunc {
	return func(bk *breaker) {
		bk.maxReadLimit = length
	}
}

// WithMaxMessagePoolLength returns a function that sets a pool size, which how many data sends with nonblocking.
func WithMaxMessagePoolLength(length int64) OptionFunc {
	return func(bk *breaker) {
		bk.broadcast = make(chan Message, length)
	}
}
