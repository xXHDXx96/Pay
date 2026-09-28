package ledger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"alipay-payment/internal/model"
	"alipay-payment/internal/store"
	"github.com/shopspring/decimal"
)

type Service struct {
	store store.Store
}

func NewService(s store.Store) *Service {
	return &Service{store: s}
}

// Transfer 执行复式记账：从 from 账户借记，向 to 账户贷记
func (s *Service) Transfer(ctx context.Context, fromType, toType, referenceType, referenceID, description string, amount decimal.Decimal, operatorID uint, traceID string) error {
	if amount.LessThanOrEqual(decimal.NewFromFloat(0)) {
		return errors.New("amount must be positive")
	}

	from, err := s.store.GetAccountByType(ctx, fromType)
	if err != nil {
		return fmt.Errorf("from account %s: %w", fromType, err)
	}
	to, err := s.store.GetAccountByType(ctx, toType)
	if err != nil {
		return fmt.Errorf("to account %s: %w", toType, err)
	}

	// 借记
	if from.BalanceCNY.LessThan(amount) {
		return fmt.Errorf("insufficient balance in account %s: have %s, need %s", fromType, from.BalanceCNY, amount)
	}
	from.BalanceCNY = from.BalanceCNY.Sub(amount)
	if err := s.store.UpdateAccount(ctx, from); err != nil {
		return err
	}
	if err := s.store.CreateLedgerEntry(ctx, &model.LedgerEntry{
		AccountID: from.ID, Direction: "DEBIT", AmountCNY: amount,
		BalanceAfter: from.BalanceCNY, ReferenceType: referenceType,
		ReferenceID: referenceID, Description: description,
		OperatorID: operatorID, TraceID: traceID,
	}); err != nil {
		return err
	}

	// 贷记
	to.BalanceCNY = to.BalanceCNY.Add(amount)
	if err := s.store.UpdateAccount(ctx, to); err != nil {
		return err
	}
	if err := s.store.CreateLedgerEntry(ctx, &model.LedgerEntry{
		AccountID: to.ID, Direction: "CREDIT", AmountCNY: amount,
		BalanceAfter: to.BalanceCNY, ReferenceType: referenceType,
		ReferenceID: referenceID, Description: description,
		OperatorID: operatorID, TraceID: traceID,
	}); err != nil {
		return err
	}

	return nil
}

// Freeze 冻结账户金额
func (s *Service) Freeze(ctx context.Context, accountType string, amount decimal.Decimal) error {
	a, err := s.store.GetAccountByType(ctx, accountType)
	if err != nil {
		return err
	}
	a.FrozenCNY = a.FrozenCNY.Add(amount)
	a.ReservedCNY = a.ReservedCNY.Add(amount)
	return s.store.UpdateAccount(ctx, a)
}

// Unfreeze 解冻
func (s *Service) Unfreeze(ctx context.Context, accountType string, amount decimal.Decimal) error {
	a, err := s.store.GetAccountByType(ctx, accountType)
	if err != nil {
		return err
	}
	a.FrozenCNY = a.FrozenCNY.Sub(amount)
	a.ReservedCNY = a.ReservedCNY.Sub(amount)
	return s.store.UpdateAccount(ctx, a)
}

// Balance 查询账户余额
func (s *Service) Balance(ctx context.Context, accountType string) (decimal.Decimal, error) {
	a, err := s.store.GetAccountByType(ctx, accountType)
	if err != nil {
		return decimal.Zero, err
	}
	return a.BalanceCNY, nil
}

// Entries 查询流水
func (s *Service) Entries(ctx context.Context, referenceID string) ([]model.LedgerEntry, error) {
	return s.store.ListLedgerEntries(ctx, referenceID)
}

var _ = time.Now
