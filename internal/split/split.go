package split

import (
	"context"
	"fmt"
	"time"

	"alipay-payment/internal/ledger"
	"alipay-payment/internal/model"
	"alipay-payment/internal/store"
	"github.com/shopspring/decimal"
)

type Service struct {
	store  store.Store
	ledger *ledger.Service
}

func NewService(s store.Store, l *ledger.Service) *Service {
	return &Service{store: s, ledger: l}
}

type SplitRule struct {
	AccountID   uint            `json:"account_id"`
	AccountType string          `json:"account_type"`
	AmountCNY   decimal.Decimal `json:"amount_cny"`
	Percentage  decimal.Decimal `json:"percentage,omitempty"`
	Description string          `json:"description"`
}

func (s *Service) ExecuteSplit(ctx context.Context, orderID uint, rules []SplitRule, operatorID uint) ([]model.SplitRecord, error) {
	order, err := s.store.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	var records []model.SplitRecord
	for _, rule := range rules {
		amount := rule.AmountCNY
		if rule.Percentage.GreaterThan(decimal.Zero) {
			amount = order.AmountCNY.Mul(rule.Percentage).Round(2)
		}

		record := &model.SplitRecord{
			OrderID:      orderID,
			AccountID:    rule.AccountID,
			AccountType:  rule.AccountType,
			AmountCNY:    amount,
			NetAmountCNY: amount,
			Status:       "PENDING",
		}
		if err := s.store.CreateSplitRecord(ctx, record); err != nil {
			return nil, err
		}

		// 执行分账转账
		if err := s.ledger.Transfer(ctx, "MERCHANT_SETTLEMENT", rule.AccountType,
			"SPLIT", fmt.Sprintf("order-%d-split-%d", orderID, rule.AccountID),
			rule.Description, amount, operatorID, fmt.Sprintf("split-%d", record.ID)); err != nil {
			record.Status = "FAILED"
			record.FailedReason = err.Error()
			_ = s.store.UpdateSplitRecord(ctx, record)
			return records, err
		}

		record.Status = "SUCCESS"
		now := time.Now()
		record.CompletedAt = &now
		_ = s.store.UpdateSplitRecord(ctx, record)
		records = append(records, *record)
	}

	return records, nil
}

func (s *Service) GetSplit(ctx context.Context, id uint) (*model.SplitRecord, error) {
	return s.store.GetSplitRecord(ctx, id)
}

func (s *Service) ListSplits(ctx context.Context, orderID uint) ([]model.SplitRecord, error) {
	return s.store.ListSplitRecords(ctx, orderID)
}
