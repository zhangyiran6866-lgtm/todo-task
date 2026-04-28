package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

var (
	auditMu     sync.RWMutex
	auditLogger = zap.NewNop()
)

func setAuditLogger(l *zap.Logger) {
	if l == nil {
		l = zap.NewNop()
	}

	auditMu.Lock()
	defer auditMu.Unlock()
	auditLogger = l
}

func getAuditLogger() *zap.Logger {
	auditMu.RLock()
	defer auditMu.RUnlock()
	return auditLogger
}

// Audit writes logs to audit channel with context tracing fields.
func Audit(ctx context.Context, module, action, msg string, fields ...zap.Field) {
	allFields := make([]zap.Field, 0, len(fields)+2)
	allFields = append(allFields, FieldsFromContext(ctx)...)
	allFields = append(allFields, zap.String(FieldModule, module))
	allFields = append(allFields, zap.String(FieldAction, action))
	allFields = append(allFields, fields...)

	getAuditLogger().Info(msg, allFields...)
}
