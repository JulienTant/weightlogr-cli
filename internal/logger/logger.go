package logger

import (
	"context"
	"log/slog"
)

type ctxKeyLogger struct{}

// WithContext stores a logger in the context.
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger{}, l)
}

// FromContext retrieves the logger from context, falling back to slog.Default.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKeyLogger{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
