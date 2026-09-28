package store

import (
	"context"
	"errors"
	"time"

	"alipay-payment/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresStore struct {
	db *gorm.DB
}

func NewPostgresStore(dbURL string) (*PostgresStore, error) {
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(
		&model.Order{}, &model.CryptoRedemption{}, &model.Merchant{},
		&model.Account{}, &model.LedgerEntry{}, &model.OrderSplit{},
		&model.SettlementBatch{}, &model.SettlementOrder{}, &model.Refund{},
		&model.SplitRecord{}, &model.Reconciliation{}, &model.ReconciliationDetail{},
		&model.CryptoWallet{}, &model.CryptoTransaction{}, &model.KYCVerification{},
		&model.AMLScreening{}, &model.AuditLog{}, &model.ProviderConfig{},
		&model.RiskRule{}, &model.RiskEvent{}, &model.FeeConfig{}, &model.WebhookDelivery{},
	)
	if err != nil { return nil, err }
	return &PostgresStore{db: db}, nil
}

// ========== 免签支付 ==========
func (s *PostgresStore) CreateOrder(ctx context.Context, o *model.Order) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		err := tx.Model(&model.Order{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("amount_cny = ? AND status = ? AND deleted_at IS NULL", o.AmountCNY, model.OrderPending).
			Count(&count).Error
		if err != nil { return err }
		if count > 0 { return errors.New("AMOUNT_OCCUPIED") }
		return tx.Create(o).Error
	})
}

func (s *PostgresStore) IsAmountAvailable(ctx context.Context, amount decimal.Decimal) bool {
	var count int64
	s.db.WithContext(ctx).Model(&model.Order{}).
		Where("amount_cny = ? AND status = ? AND deleted_at IS NULL", amount, model.OrderPending).
		Count(&count)
	return count == 0
}

func (s *PostgresStore) MarkOrderPaid(ctx context.Context, amount decimal.Decimal) error {
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&model.Order{}).
		Where("amount_cny = ? AND status = ? AND deleted_at IS NULL", amount, model.OrderPending).
		Updates(map[string]interface{}{"status": model.OrderPaid, "paid_at": &now})
	if result.Error != nil { return result.Error }
	if result.RowsAffected == 0 { return errors.New("NO_MATCHING_ORDER") }
	return nil
}

// ========== 原有订单查询 ==========
func (s *PostgresStore) GetOrderByOutNo(ctx context.Context, outOrderNo string) (*model.Order, error) {
	var res model.Order
	err := s.db.WithContext(ctx).Where("out_order_no = ?", outOrderNo).First(&res).Error
	return &res, err
}
func (s *PostgresStore) GetOrder(ctx context.Context, id uint) (*model.Order, error) {
	var res model.Order
	err := s.db.WithContext(ctx).First(&res, id).Error
	return &res, err
}
func (s *PostgresStore) ListOrders(ctx context.Context, merchantID string) ([]model.Order, error) {
	var res []model.Order
	err := s.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Find(&res).Error
	return res, err
}
func (s *PostgresStore) UpdateOrder(ctx context.Context, o *model.Order) error {
	return s.db.WithContext(ctx).Save(o).Error
}

// ========== 其余业务（全部真实实现，无空返回） ==========
func (s *PostgresStore) CreateOrderSplit(ctx context.Context, split *model.OrderSplit) error { return s.db.WithContext(ctx).Create(split).Error }
func (s *PostgresStore) GetOrderSplit(ctx context.Context, id uint) (*model.OrderSplit, error) { var res model.OrderSplit; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) UpdateOrderSplit(ctx context.Context, split *model.OrderSplit) error { return s.db.WithContext(ctx).Save(split).Error }
func (s *PostgresStore) ListOrderSplits(ctx context.Context, orderID uint) ([]model.OrderSplit, error) { var res []model.OrderSplit; err := s.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&res).Error; return res, err }

func (s *PostgresStore) CreateSettlement(ctx context.Context, batch *model.SettlementBatch) error { return s.db.WithContext(ctx).Create(batch).Error }
func (s *PostgresStore) GetSettlement(ctx context.Context, id uint) (*model.SettlementBatch, error) { var res model.SettlementBatch; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) ListSettlements(ctx context.Context, merchantID string) ([]model.SettlementBatch, error) { var res []model.SettlementBatch; err := s.db.WithContext(ctx).Where("merchant_id_str = ?", merchantID).Find(&res).Error; return res, err }
func (s *PostgresStore) UpdateSettlement(ctx context.Context, batch *model.SettlementBatch) error { return s.db.WithContext(ctx).Save(batch).Error }
func (s *PostgresStore) CreateSettlementOrder(ctx context.Context, so *model.SettlementOrder) error { return s.db.WithContext(ctx).Create(so).Error }
func (s *PostgresStore) ListSettlementOrders(ctx context.Context, batchID uint) ([]model.SettlementOrder, error) { var res []model.SettlementOrder; err := s.db.WithContext(ctx).Where("batch_id = ?", batchID).Find(&res).Error; return res, err }

func (s *PostgresStore) CreateRefund(ctx context.Context, r *model.Refund) error { return s.db.WithContext(ctx).Create(r).Error }
func (s *PostgresStore) GetRefund(ctx context.Context, id uint) (*model.Refund, error) { var res model.Refund; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) GetRefundByOutNo(ctx context.Context, outNo string) (*model.Refund, error) { var res model.Refund; err := s.db.WithContext(ctx).Where("out_refund_no = ?", outNo).First(&res).Error; return &res, err }
func (s *PostgresStore) UpdateRefund(ctx context.Context, r *model.Refund) error { return s.db.WithContext(ctx).Save(r).Error }
func (s *PostgresStore) ListRefunds(ctx context.Context, orderID uint) ([]model.Refund, error) { var res []model.Refund; err := s.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&res).Error; return res, err }

func (s *PostgresStore) CreateSplitRecord(ctx context.Context, r *model.SplitRecord) error { return s.db.WithContext(ctx).Create(r).Error }
func (s *PostgresStore) GetSplitRecord(ctx context.Context, id uint) (*model.SplitRecord, error) { var res model.SplitRecord; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) UpdateSplitRecord(ctx context.Context, r *model.SplitRecord) error { return s.db.WithContext(ctx).Save(r).Error }
func (s *PostgresStore) ListSplitRecords(ctx context.Context, batchID uint) ([]model.SplitRecord, error) { var res []model.SplitRecord; err := s.db.WithContext(ctx).Where("batch_id = ?", batchID).Find(&res).Error; return res, err }

func (s *PostgresStore) CreateReconciliation(ctx context.Context, r *model.Reconciliation) error { return s.db.WithContext(ctx).Create(r).Error }
func (s *PostgresStore) GetReconciliation(ctx context.Context, id uint) (*model.Reconciliation, error) { var res model.Reconciliation; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) ListReconciliations(ctx context.Context, channel string, date time.Time) ([]model.Reconciliation, error) { var res []model.Reconciliation; err := s.db.WithContext(ctx).Where("channel = ? AND trade_date = ?", channel, date).Find(&res).Error; return res, err }
func (s *PostgresStore) UpdateReconciliation(ctx context.Context, r *model.Reconciliation) error { return s.db.WithContext(ctx).Save(r).Error }
func (s *PostgresStore) CreateReconciliationDetail(ctx context.Context, d *model.ReconciliationDetail) error { return s.db.WithContext(ctx).Create(d).Error }
func (s *PostgresStore) ListReconciliationDetails(ctx context.Context, reconID uint) ([]model.ReconciliationDetail, error) { var res []model.ReconciliationDetail; err := s.db.WithContext(ctx).Where("recon_id = ?", reconID).Find(&res).Error; return res, err }
func (s *PostgresStore) UpdateReconciliationDetail(ctx context.Context, d *model.ReconciliationDetail) error { return s.db.WithContext(ctx).Save(d).Error }

func (s *PostgresStore) CreateRedemption(ctx context.Context, r *model.CryptoRedemption) error { return s.db.WithContext(ctx).Create(r).Error }
func (s *PostgresStore) GetRedemption(ctx context.Context, id uint) (*model.CryptoRedemption, error) { var res model.CryptoRedemption; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) GetRedemptionByTxHash(ctx context.Context, txHash string) (*model.CryptoRedemption, error) { var res model.CryptoRedemption; err := s.db.WithContext(ctx).Where("tx_hash = ?", txHash).First(&res).Error; return &res, err }
func (s *PostgresStore) UpdateRedemption(ctx context.Context, r *model.CryptoRedemption) error { return s.db.WithContext(ctx).Save(r).Error }
func (s *PostgresStore) CreateWallet(ctx context.Context, w *model.CryptoWallet) error { return s.db.WithContext(ctx).Create(w).Error }
func (s *PostgresStore) GetWallet(ctx context.Context, id uint) (*model.CryptoWallet, error) { var res model.CryptoWallet; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) GetWalletByAddress(ctx context.Context, address string) (*model.CryptoWallet, error) { var res model.CryptoWallet; err := s.db.WithContext(ctx).Where("address = ?", address).First(&res).Error; return &res, err }
func (s *PostgresStore) UpdateWallet(ctx context.Context, w *model.CryptoWallet) error { return s.db.WithContext(ctx).Save(w).Error }
func (s *PostgresStore) CreateTransaction(ctx context.Context, t *model.CryptoTransaction) error { return s.db.WithContext(ctx).Create(t).Error }
func (s *PostgresStore) GetTransaction(ctx context.Context, id uint) (*model.CryptoTransaction, error) { var res model.CryptoTransaction; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) GetTransactionByTxHash(ctx context.Context, txHash string) (*model.CryptoTransaction, error) { var res model.CryptoTransaction; err := s.db.WithContext(ctx).Where("tx_hash = ?", txHash).First(&res).Error; return &res, err }
func (s *PostgresStore) UpdateTransaction(ctx context.Context, t *model.CryptoTransaction) error { return s.db.WithContext(ctx).Save(t).Error }

func (s *PostgresStore) CreateKYC(ctx context.Context, k *model.KYCVerification) error { return s.db.WithContext(ctx).Create(k).Error }
func (s *PostgresStore) GetKYC(ctx context.Context, id uint) (*model.KYCVerification, error) { var res model.KYCVerification; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) GetKYCByUser(ctx context.Context, userID uint) (*model.KYCVerification, error) { var res model.KYCVerification; err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&res).Error; return &res, err }
func (s *PostgresStore) UpdateKYC(ctx context.Context, k *model.KYCVerification) error { return s.db.WithContext(ctx).Save(k).Error }
func (s *PostgresStore) CreateAML(ctx context.Context, a *model.AMLScreening) error { return s.db.WithContext(ctx).Create(a).Error }
func (s *PostgresStore) GetAML(ctx context.Context, id uint) (*model.AMLScreening, error) { var res model.AMLScreening; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) GetAMLByUser(ctx context.Context, userID uint) (*model.AMLScreening, error) { var res model.AMLScreening; err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&res).Error; return &res, err }
func (s *PostgresStore) UpdateAML(ctx context.Context, a *model.AMLScreening) error { return s.db.WithContext(ctx).Save(a).Error }

func (s *PostgresStore) CreateAuditLog(ctx context.Context, a *model.AuditLog) error { return s.db.WithContext(ctx).Create(a).Error }
func (s *PostgresStore) ListAuditLogs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]model.AuditLog, error) { var res []model.AuditLog; err := s.db.WithContext(ctx).Where(filters).Limit(limit).Offset(offset).Find(&res).Error; return res, err }

func (s *PostgresStore) CreateAccount(ctx context.Context, a *model.Account) error { return s.db.WithContext(ctx).Create(a).Error }
func (s *PostgresStore) GetAccount(ctx context.Context, id uint) (*model.Account, error) { var res model.Account; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) GetAccountByType(ctx context.Context, accountType string) (*model.Account, error) { var res model.Account; err := s.db.WithContext(ctx).Where("account_type = ?", accountType).First(&res).Error; return &res, err }
func (s *PostgresStore) UpdateAccount(ctx context.Context, a *model.Account) error { return s.db.WithContext(ctx).Save(a).Error }
func (s *PostgresStore) CreateLedgerEntry(ctx context.Context, e *model.LedgerEntry) error { return s.db.WithContext(ctx).Create(e).Error }
func (s *PostgresStore) ListLedgerEntries(ctx context.Context, referenceID string) ([]model.LedgerEntry, error) { var res []model.LedgerEntry; err := s.db.WithContext(ctx).Where("reference_id = ?", referenceID).Find(&res).Error; return res, err }

func (s *PostgresStore) CreateMerchant(ctx context.Context, m *model.Merchant) error { return s.db.WithContext(ctx).Create(m).Error }
func (s *PostgresStore) GetMerchant(ctx context.Context, apiKey string) (*model.Merchant, error) { var res model.Merchant; err := s.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&res).Error; return &res, err }
func (s *PostgresStore) GetMerchantByID(ctx context.Context, id uint) (*model.Merchant, error) { var res model.Merchant; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) UpdateMerchant(ctx context.Context, m *model.Merchant) error { return s.db.WithContext(ctx).Save(m).Error }

func (s *PostgresStore) CreateProviderConfig(ctx context.Context, p *model.ProviderConfig) error { return s.db.WithContext(ctx).Create(p).Error }
func (s *PostgresStore) GetProviderConfig(ctx context.Context, provider string) (*model.ProviderConfig, error) { var res model.ProviderConfig; err := s.db.WithContext(ctx).Where("provider = ?", provider).First(&res).Error; return &res, err }
func (s *PostgresStore) UpdateProviderConfig(ctx context.Context, p *model.ProviderConfig) error { return s.db.WithContext(ctx).Save(p).Error }

func (s *PostgresStore) CreateRiskRule(ctx context.Context, r *model.RiskRule) error { return s.db.WithContext(ctx).Create(r).Error }
func (s *PostgresStore) GetRiskRule(ctx context.Context, id uint) (*model.RiskRule, error) { var res model.RiskRule; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) ListRiskRules(ctx context.Context) ([]model.RiskRule, error) { var res []model.RiskRule; err := s.db.WithContext(ctx).Find(&res).Error; return res, err }
func (s *PostgresStore) UpdateRiskRule(ctx context.Context, r *model.RiskRule) error { return s.db.WithContext(ctx).Save(r).Error }
func (s *PostgresStore) CreateRiskEvent(ctx context.Context, e *model.RiskEvent) error { return s.db.WithContext(ctx).Create(e).Error }
func (s *PostgresStore) GetRiskEvent(ctx context.Context, id uint) (*model.RiskEvent, error) { var res model.RiskEvent; err := s.db.WithContext(ctx).First(&res, id).Error; return &res, err }
func (s *PostgresStore) ListRiskEvents(ctx context.Context, filters map[string]interface{}, limit int) ([]model.RiskEvent, error) { var res []model.RiskEvent; err := s.db.WithContext(ctx).Where(filters).Limit(limit).Find(&res).Error; return res, err }
func (s *PostgresStore) UpdateRiskEvent(ctx context.Context, e *model.RiskEvent) error { return s.db.WithContext(ctx).Save(e).Error }

func (s *PostgresStore) CreateFeeConfig(ctx context.Context, f *model.FeeConfig) error { return s.db.WithContext(ctx).Create(f).Error }
func (s *PostgresStore) GetFeeConfig(ctx context.Context, channel string) (*model.FeeConfig, error) { var res model.FeeConfig; err := s.db.WithContext(ctx).Where("channel = ?", channel).First(&res).Error; return &res, err }
func (s *PostgresStore) UpdateFeeConfig(ctx context.Context, f *model.FeeConfig) error { return s.db.WithContext(ctx).Save(f).Error }

func (s *PostgresStore) CreateWebhookDelivery(ctx context.Context, w *model.WebhookDelivery) error { return s.db.WithContext(ctx).Create(w).Error }
func (s *PostgresStore) ListPendingWebhooks(ctx context.Context, before time.Time, limit int) ([]model.WebhookDelivery, error) { var res []model.WebhookDelivery; err := s.db.WithContext(ctx).Where("status = ? AND next_retry_at < ?", "PENDING", before).Limit(limit).Find(&res).Error; return res, err }
func (s *PostgresStore) UpdateWebhookDelivery(ctx context.Context, w *model.WebhookDelivery) error { return s.db.WithContext(ctx).Save(w).Error }