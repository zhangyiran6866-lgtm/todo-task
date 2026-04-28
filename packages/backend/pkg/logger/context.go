package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const (
	requestIDContextKey contextKey = "request_id"
	userIDContextKey    contextKey = "user_id"
)

const (
	FieldRequestID = "request_id"
	FieldUserID    = "user_id"
	FieldModule    = "module"
	FieldAction    = "action"
)

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey).(string)
	return requestID
}

func WithUserID(ctx context.Context, userID string) context.Context {
	if userID == "" {
		return ctx
	}
	return context.WithValue(ctx, userIDContextKey, userID)
}

func UserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(userIDContextKey).(string)
	return userID
}

func FieldsFromContext(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0, 2)

	if requestID := RequestIDFromContext(ctx); requestID != "" {
		fields = append(fields, zap.String(FieldRequestID, requestID))
	}
	if userID := UserIDFromContext(ctx); userID != "" {
		fields = append(fields, zap.String(FieldUserID, userID))
	}

	return fields
}

func WithContext(base *zap.Logger, ctx context.Context) *zap.Logger {
	if base == nil {
		base = zap.NewNop()
	}
	return base.With(FieldsFromContext(ctx)...)
}
