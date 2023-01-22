package ctxutils

import (
	"context"
	"time"
)

func WithNoLimit(ctx context.Context) context.Context {
	return unlimitedContext{
		Wrapped: ctx,
		done:    make(chan struct{}),
	}
}

type unlimitedContext struct {
	Wrapped context.Context
	done    chan struct{}
}

func (unlimitedContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (ctx unlimitedContext) Done() <-chan struct{} {
	return ctx.done
}

func (unlimitedContext) Err() error {
	return nil
}

func (ctx unlimitedContext) Value(key interface{}) interface{} {
	return ctx.Wrapped.Value(key)
}
