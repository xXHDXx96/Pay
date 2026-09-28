package settlement

import (
	"context"
	"errors"
	"fmt"
	"time"

	"alipay-payment/internal/ledger"
	"alipay-payment/internal/model"
	"alipay-payment/internal/store"
	"github.com/shopspring/decimal"
)

type BankAPI interface {
	Payout(ctx context.Context, req PayoutRequest) (string, error)
	QueryPayout(ctx context.Context, seqNo string) (*PayoutResult, error)
}

type PayoutRequest struct {
	BankCode    string
	AccountName string
	AccountNo   string
	AmountCNY   decimal.Decimal
	SeqNo       string
	Description string
}

type PayoutResult struct {
	SeqNo     string
	Status    string
	ReceiptNo string
	FailedMsg string
}

type Service struct {
	store   store.Store
	ledger  *ledger.Service
	bankAPI BankAPI
	runner  *Runner
	stopCh  chan struct{}
}

func NewService(s store.Store, l *ledger.Service, bank BankAPI) *Service {
	svc := &Service{store: s, ledger: l, bankAPI: bank, stopCh: make(chan struct{})}
	svc.runner = &Runner{svc: svc, interval: 24 * time.Hour, stopCh: make(chan struct{})}
	return svc
}

func (s *Service) Runner() *Runner { return s.runner }

func (s *Service) CreateBatch(ctx context.Context, merchantID uint) (*model.SettlementBatch, error) {
	orders, err := s.store.ListOrders(ctx, fmt.Sprintf("%d", merchantID))
	if err != nil {
		return nil, err
	}
	var total decimal.Decimal
	var count int
	var orderIDs []uint
	for _, o := range orders {
		if o.Status == "PAID" && o.SettledAt == nil {
			total = total.Add(o.AmountCNY)
			count++
			orderIDs = append(orderIDs, o.ID)
		}
	}
	if count == 0 {
		return nil, errors.New("no orders to settle")
	}

	feeRate := decimal.NewFromFloat(0.006)
	fee := total.Mul(feeRate).Round(2)
	net := total.Sub(fee)

	batch := &model.SettlementBatch{
		MerchantID:     merchantID,
		BatchNo:        fmt.Sprintf("S%s_%d", time.Now().Format("20060102"), time.Now().Unix()),
		TotalAmountCNY: total,
		OrderCount:     count,
		FeeCNY:         fee,
		NetAmountCNY:   net,
		Status:         "PENDING",
	}
	if err := s.store.CreateSettlement(ctx, batch); err != nil {
		return nil, err
	}

	for _, oid := range orderIDs {
		_ = s.store.CreateSettlementOrder(ctx, &model.SettlementOrder{
			BatchID: batch.ID, OrderID: oid, AmountCNY: decimal.Zero,
			FeeCNY: decimal.Zero, NetAmountCNY: decimal.Zero,
		})
	}

	// 分账：平台手续费 + 商户结算
	if err := s.ledger.Transfer(ctx, "MERCHANT_SETTLEMENT", "PLATFORM_FEE",
		"SETTLEMENT_FEE", batch.BatchNo, "平台服务费", fee, 0, fmt.Sprintf("batch-%d", batch.ID)); err != nil {
		return nil, err
	}

	batch.Status = "PROCESSING"
	_ = s.store.UpdateSettlement(ctx, batch)
	return batch, nil
}

func (s *Service) Confirm(ctx context.Context, batchID uint, operatorID uint) (*model.SettlementBatch, error) {
	batch, err := s.store.GetSettlement(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if batch.Status != "PROCESSING" {
		return nil, errors.New("batch not in PROCESSING status")
	}

	resp, err := s.bankAPI.Payout(ctx, PayoutRequest{
		BankCode:    batch.BankCode,
		AccountName: batch.BankAccountName,
		AccountNo:   batch.BankAccountNo,
		AmountCNY:   batch.NetAmountCNY,
		SeqNo:       batch.BatchNo,
		Description: fmt.Sprintf("结算批次 %s", batch.BatchNo),
	})
	if err != nil {
		batch.Status = "FAILED"
		batch.FailedReason = err.Error()
		batch.RetryCount++
		_ = s.store.UpdateSettlement(ctx, batch)
		return nil, err
	}

	batch.BankSeqNo = resp
	batch.Status = "PAID"
	now := time.Now()
	batch.PaidAt = &now
	batch.ConfirmedBy = operatorID
	batch.ConfirmedAt = &now
	if err := s.store.UpdateSettlement(ctx, batch); err != nil {
		return nil, err
	}

	// 商户结算账户入账
	_ = s.ledger.Transfer(ctx, "PLATFORM_FEE", "MERCHANT_SETTLEMENT",
		"SETTLEMENT_PAYOUT", batch.BatchNo, "商户结算打款", batch.NetAmountCNY, operatorID, fmt.Sprintf("batch-%d", batch.ID))

	return batch, nil
}

func (s *Service) Retry(ctx context.Context, batchID uint) (*model.SettlementBatch, error) {
	batch, err := s.store.GetSettlement(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if batch.Status != "FAILED" {
		return nil, errors.New("batch not in FAILED status")
	}
	if batch.RetryCount >= 5 {
		return nil, errors.New("max retries exceeded")
	}
	batch.Status = "PROCESSING"
	_ = s.store.UpdateSettlement(ctx, batch)
	return s.Confirm(ctx, batchID, batch.ConfirmedBy)
}

type Runner struct {
	svc      *Service
	interval time.Duration
	stopCh   chan struct{}
}

func (r *Runner) Start() { go r.run() }
func (r *Runner) Stop()  { close(r.stopCh) }
func (r *Runner) Trigger() {
	select {
	case r.svc.stopCh <- struct{}{}:
	default:
	}
}
func (r *Runner) run() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			_ = r.svc.runBatch()
		}
	}
}
func (s *Service) runBatch() error {
	ctx := context.Background()
	// 简化：遍历所有商户
	for mid := uint(1); mid <= 10; mid++ {
		_, _ = s.CreateBatch(ctx, mid)
	}
	return nil
}
