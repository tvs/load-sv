package log

import (
	"context"
	"log/slog"
)

type contextLogger struct{}

func ContextWithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, contextLogger{}, l)
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(contextLogger{}).(*slog.Logger); ok {
		return l
	}

	return slog.Default()
}
