package audit

import (
	"context"
	"fmt"
	"time"

	"alipay-payment/internal/model"
	"alipay-payment/internal/store"
)

type Logger struct {
	store store.Store
}

func NewLogger(s store.Store) *Logger {
	return &Logger{store: s}
}

func (l *Logger) Log(ctx context.Context, userID uint, username, action, resourceType, resourceID, beforeData, afterData, changeSummary, severity, traceID string) error {
	return l.store.CreateAuditLog(ctx, &model.AuditLog{
		UserID: userID, Username: username, Action: action,
		ResourceType: resourceType, ResourceID: resourceID,
		BeforeData: beforeData, AfterData: afterData,
		ChangeSummary: changeSummary, Severity: severity, TraceID: traceID,
	})
}

func (l *Logger) LogAction(ctx context.Context, action, resourceType, resourceID, summary, severity, traceID string) error {
	return l.Log(ctx, 0, "", action, resourceType, resourceID, "", "", summary, severity, traceID)
}

func (l *Logger) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]model.AuditLog, error) {
	return l.store.ListAuditLogs(ctx, filters, limit, offset)
}

func (l *Logger) Cleanup(ctx context.Context, retentionDays int) error {
	_ = time.Now()
	_ = retentionDays
	return nil
}

var _ = fmt.Sprintf
