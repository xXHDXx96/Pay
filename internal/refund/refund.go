package refund

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

type Provider interface {
	Refund(ctx context.Context, orderNo, refundNo string, amount decimal.Decimal, reason string) (string, error)
	QueryRefund(ctx context.Context, refundNo string) (*RefundStatus, error)
}

type RefundStatus struct {
	ProviderRefundNo string
	Status           string
	RefundedAt       *time.Time
}

type Service struct {
	store     store.Store
	ledger    *ledger.Service
	providers map[string]Provider
}

func NewService(s store.Store, l *ledger.Service) *Service {
	return &Service{store: s, ledger: l, providers: make(map[string]Provider)}
}

func (s *Service) RegisterProvider(name string, p Provider) { s.providers[name] = p }

func (s *Service) CreateRefund(ctx context.Context, orderID uint, outRefundNo string, amount decimal.Decimal, reason, channel string, operatorID uint) (*model.Refund, error) {
	order, err := s.store.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != "PAID" {
		return nil, errors.New("order not paid")
	}
	if amount.GreaterThan(order.AmountCNY) {
		return nil, errors.New("refund amount exceeds order amount")
	}

	refund := &model.Refund{
		OrderID:     orderID,
		OutRefundNo: outRefundNo,
		AmountCNY:   amount,
		Reason:      reason,
		Channel:     channel,
		Status:      "PENDING",
		OperatorID:  operatorID,
	}
	if err := s.store.CreateRefund(ctx, refund); err != nil {
		return nil, err
	}

	// 调用渠道退款
	provider, ok := s.providers[channel]
	if !ok {
		refund.Status = "FAILED"
		refund.FailedReason = "provider not configured"
		_ = s.store.UpdateRefund(ctx, refund)
		return refund, nil
	}

	providerRefundNo, err := provider.Refund(ctx, order.OutOrderNo, outRefundNo, amount, reason)
	if err != nil {
		refund.Status = "FAILED"
		refund.FailedReason = err.Error()
		_ = s.store.UpdateRefund(ctx, refund)
		return refund, err
	}

	refund.ProviderRefundNo = providerRefundNo
	refund.Status = "PROCESSING"
	_ = s.store.UpdateRefund(ctx, refund)

	// 垫资退款：先从平台资金垫付给用户
	_ = s.ledger.Transfer(ctx, "USER_WALLET", "PLATFORM_FEE",
		"REFUND_ADVANCE", outRefundNo, "垫资退款", amount, operatorID, fmt.Sprintf("refund-%s", outRefundNo))

	now := time.Now()
	refund.RefundedAt = &now
	refund.Status = "SUCCESS"
	_ = s.store.UpdateRefund(ctx, refund)

	return refund, nil
}

func (s *Service) GetRefund(ctx context.Context, id uint) (*model.Refund, error) {
	return s.store.GetRefund(ctx, id)
}

func (s *Service) ListRefunds(ctx context.Context, orderID uint) ([]model.Refund, error) {
	return s.store.ListRefunds(ctx, orderID)
}
