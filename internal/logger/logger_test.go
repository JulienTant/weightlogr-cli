package logger

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromContext(t *testing.T) {
	t.Run("no logger returns default", func(t *testing.T) {
		got := FromContext(context.Background())
		assert.Equal(t, slog.Default(), got)
	})

	t.Run("wrong type returns default", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), ctxKeyLogger{}, "not-a-logger")
		got := FromContext(ctx)
		assert.Equal(t, slog.Default(), got)
	})

	t.Run("nil context panics", func(t *testing.T) {
		assert.Panics(t, func() {
			FromContext(nil) //nolint:staticcheck
		})
	})

	t.Run("returned logger is usable", func(t *testing.T) {
		tests := []struct {
			name string
			ctx  context.Context
		}{
			{"default logger", context.Background()},
			{"custom logger", WithContext(context.Background(), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				l := FromContext(tt.ctx)
				assert.NotPanics(t, func() {
					l.Info("test", "key", "value")
					l.Warn("warning", "count", 42)
					l.Error("error", "err", "broke")
				})
			})
		}
	})

	t.Run("custom logger actually writes", func(t *testing.T) {
		var buf bytes.Buffer
		l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
		ctx := WithContext(context.Background(), l)

		FromContext(ctx).Info("hello", "who", "world")

		output := buf.String()
		require.NotEmpty(t, output)
		assert.Contains(t, output, "hello")
		assert.Contains(t, output, "who")
		assert.Contains(t, output, "world")
	})
}

func TestWithContext(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		l := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		ctx := WithContext(context.Background(), l)
		assert.Same(t, l, FromContext(ctx))
	})

	t.Run("overwrites existing", func(t *testing.T) {
		first := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		second := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))

		ctx := WithContext(context.Background(), first)
		ctx = WithContext(ctx, second)

		assert.Same(t, second, FromContext(ctx))
		assert.NotSame(t, first, FromContext(ctx))
	})

	t.Run("nil context panics", func(t *testing.T) {
		l := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		assert.Panics(t, func() {
			WithContext(nil, l) //nolint:staticcheck
		})
	})

	t.Run("preserves parent values", func(t *testing.T) {
		type otherKey struct{}
		parent := context.WithValue(context.Background(), otherKey{}, "preserved")
		l := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

		ctx := WithContext(parent, l)

		assert.Equal(t, "preserved", ctx.Value(otherKey{}))
		assert.Same(t, l, FromContext(ctx))
	})
}
