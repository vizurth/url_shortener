package logger

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey struct{}
type requestIDKey struct{}

type Logger struct {
	z *zap.Logger
}

func New() (*Logger, error) {
	z, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &Logger{z: z}, nil
}

func With(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

func From(ctx context.Context) *Logger {
	if l, ok := ctx.Value(loggerKey{}).(*Logger); ok {
		return l
	}
	return nopLogger()
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.z.Debug(msg, l.appendCtx(ctx, fields)...)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.z.Info(msg, l.appendCtx(ctx, fields)...)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.z.Warn(msg, l.appendCtx(ctx, fields)...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.z.Error(msg, l.appendCtx(ctx, fields)...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	l.z.Fatal(msg, l.appendCtx(ctx, fields)...)
}

func (l *Logger) Sync() error {
	return l.z.Sync()
}

func (l *Logger) appendCtx(ctx context.Context, fields []zap.Field) []zap.Field {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok && id != "" {
		return append(fields, zap.String("requestID", id))
	}
	return fields
}

func nopLogger() *Logger {
	return &Logger{z: zap.NewNop()}
}
