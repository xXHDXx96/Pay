package compliance

import (
	"context"
	"fmt"
	"time"

	"alipay-payment/internal/audit"
	"alipay-payment/internal/model"
	"alipay-payment/internal/store"
)

type Reporter struct {
	store  store.Store
	audit  *audit.Logger
	stopCh chan struct{}
}

func NewReporter(s store.Store, a *audit.Logger) *Reporter {
	return &Reporter{store: s, audit: a, stopCh: make(chan struct{})}
}

func (r *Reporter) Start() {
	go r.run()
}
func (r *Reporter) Stop() { close(r.stopCh) }
func (r *Reporter) run() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			_ = r.DailyReport(context.Background())
		}
	}
}

func (r *Reporter) DailyReport(ctx context.Context) error {
	_ = r.audit.LogAction(ctx, "COMPLIANCE_REPORT", "REPORT", "daily",
		"生成日报并上报", "INFO", fmt.Sprintf("report-%s", time.Now().Format("20060102")))
	return nil
}

func (r *Reporter) ReportAMLAlert(ctx context.Context, alert map[string]interface{}) error {
	return r.audit.LogAction(ctx, "AML_ALERT_REPORT", "AML", fmt.Sprintf("%v", alert["user_id"]),
		"上报可疑交易", "WARN", fmt.Sprintf("aml-%s", time.Now().Format("20060102")))
}

func (r *Reporter) ReportLargeTransaction(ctx context.Context, orderID uint, amount float64) error {
	return r.audit.LogAction(ctx, "LARGE_TX_REPORT", "ORDER", fmt.Sprintf("%d", orderID),
		fmt.Sprintf("大额交易上报: %.2f CNY", amount), "WARN", fmt.Sprintf("large-%d", orderID))
}

func (r *Reporter) GenerateSTARReport(ctx context.Context) (string, error) {
	return fmt.Sprintf("STAR-REPORT-%s", time.Now().Format("20060102")), nil
}

var _ = model.AuditLog{}
