package xlog

import (
	"context"
	"time"
)

func NewWithCtx(ctx context.Context) Logger {
	return WithCtx(New(), ctx)
}

func WithCtx(xl Logger, ctx context.Context) Logger {
	if ctx == nil {
		ctx = context.Background()
	}
	l := convertLogger(xl)
	l.ctx = ctx
	return l
}

func (l *logger) Deadline() (deadline time.Time, ok bool) {
	return l.ctx.Deadline()
}

func (l *logger) Done() <-chan struct{} {
	return l.ctx.Done()
}

func (l *logger) Err() error {
	return l.ctx.Err()
}

func (l *logger) Value(key any) any {
	return l.ctx.Value(key)
}
