package store

import (
	"context"
	"time"

	"alipay-payment/internal/model"
	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGorm(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (s *GormStore) CreateOrder(ctx context.Context, o *model.Order) error {
	return s.db.WithContext(ctx).Create(o).Error
}

func (s *GormStore) GetOrder(ctx context.Context, id uint) (*model.Order, error) {
	var o model.Order
	if err := s.db.WithContext(ctx).First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *GormStore) GetOrderByOutNo(ctx context.Context, outNo string) (*model.Order, error) {
	var o model.Order
	if err := s.db.WithContext(ctx).Where("out_order_no = ?", outNo).First(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *GormStore) UpdateOrder(ctx context.Context, o *model.Order) error {
	return s.db.WithContext(ctx).Save(o).Error
}

func (s *GormStore) ListOrders(ctx context.Context, merchantID string) ([]model.Order, error) {
	var orders []model.Order
	q := s.db.WithContext(ctx)
	if merchantID != "" {
		q = q.Where("merchant_id = ?", merchantID)
	}
	if err := q.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *GormStore) CreateRedemption(ctx context.Context, r *model.CryptoRedemption) error {
	return s.db.WithContext(ctx).Create(r).Error
}

func (s *GormStore) GetRedemption(ctx context.Context, orderID uint) (*model.CryptoRedemption, error) {
	var r model.CryptoRedemption
	if err := s.db.WithContext(ctx).First(&r, orderID).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *GormStore) GetRedemptionByTxHash(ctx context.Context, txHash string) (*model.CryptoRedemption, error) {
	var r model.CryptoRedemption
	if err := s.db.WithContext(ctx).Where("tx_hash = ?", txHash).First(&r).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *GormStore) UpdateRedemption(ctx context.Context, r *model.CryptoRedemption) error {
	return s.db.WithContext(ctx).Save(r).Error
}

func (s *GormStore) CreateRefund(ctx context.Context, r *model.Refund) error {
	return s.db.WithContext(ctx).Create(r).Error
}

func (s *GormStore) GetRefund(ctx context.Context, id uint) (*model.Refund, error) {
	var r model.Refund
	if err := s.db.WithContext(ctx).First(&r, id).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *GormStore) GetRefundByOutNo(ctx context.Context, outNo string) (*model.Refund, error) {
	var r model.Refund
	if err := s.db.WithContext(ctx).Where("out_refund_no = ?", outNo).First(&r).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *GormStore) UpdateRefund(ctx context.Context, r *model.Refund) error {
	return s.db.WithContext(ctx).Save(r).Error
}

func (s *GormStore) ListRefunds(ctx context.Context, orderID uint) ([]model.Refund, error) {
	var refunds []model.Refund
	if err := s.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&refunds).Error; err != nil {
		return nil, err
	}
	return refunds, nil
}

func (s *GormStore) CreateSettlement(ctx context.Context, b *model.SettlementBatch) error {
	return s.db.WithContext(ctx).Create(b).Error
}

func (s *GormStore) GetSettlement(ctx context.Context, id uint) (*model.SettlementBatch, error) {
	var b model.SettlementBatch
	if err := s.db.WithContext(ctx).First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *GormStore) ListSettlements(ctx context.Context, merchantID string) ([]model.SettlementBatch, error) {
	var batches []model.SettlementBatch
	q := s.db.WithContext(ctx)
	if merchantID != "" {
		q = q.Where("merchant_id = ?", merchantID)
	}
	if err := q.Find(&batches).Error; err != nil {
		return nil, err
	}
	return batches, nil
}

func (s *GormStore) UpdateSettlement(ctx context.Context, b *model.SettlementBatch) error {
	return s.db.WithContext(ctx).Save(b).Error
}

func (s *GormStore) CreateWebhookDelivery(ctx context.Context, w *model.WebhookDelivery) error {
	return s.db.WithContext(ctx).Create(w).Error
}

func (s *GormStore) ListPendingWebhooks(ctx context.Context, before time.Time, limit int) ([]model.WebhookDelivery, error) {
	var ws []model.WebhookDelivery
	err := s.db.WithContext(ctx).
		Where("status = 'PENDING' AND (next_retry_at IS NULL OR next_retry_at <= ?)", before).
		Limit(limit).Find(&ws).Error
	return ws, err
}

func (s *GormStore) UpdateWebhookDelivery(ctx context.Context, w *model.WebhookDelivery) error {
	return s.db.WithContext(ctx).Save(w).Error
}

func (s *GormStore) CreateMerchant(ctx context.Context, m *model.Merchant) error {
	return s.db.WithContext(ctx).Create(m).Error
}

func (s *GormStore) GetMerchant(ctx context.Context, apiKey string) (*model.Merchant, error) {
	var m model.Merchant
	if err := s.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *GormStore) GetMerchantByID(ctx context.Context, id uint) (*model.Merchant, error) {
	var m model.Merchant
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *GormStore) UpdateMerchant(ctx context.Context, m *model.Merchant) error {
	return s.db.WithContext(ctx).Save(m).Error
}
