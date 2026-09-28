package store

import (
	"context"
	"errors"
	"time"

	"alipay-payment/internal/model"

	"github.com/shopspring/decimal"
)

type Store interface {
	// ========== 免签支付相关 ==========
	CreateOrder(ctx context.Context, o *model.Order) error
	IsAmountAvailable(ctx context.Context, amount decimal.Decimal) bool
	MarkOrderPaid(ctx context.Context, amount decimal.Decimal) error

	// ========== 原有订单相关 ==========
	GetOrderByOutNo(ctx context.Context, outOrderNo string) (*model.Order, error)
	GetOrder(ctx context.Context, id uint) (*model.Order, error)
	ListOrders(ctx context.Context, merchantID string) ([]model.Order, error)
	UpdateOrder(ctx context.Context, o *model.Order) error

	// ========== 其他业务（保持完整） ==========
	// Splits
	CreateOrderSplit(ctx context.Context, s *model.OrderSplit) error
	GetOrderSplit(ctx context.Context, id uint) (*model.OrderSplit, error)
	UpdateOrderSplit(ctx context.Context, s *model.OrderSplit) error
	ListOrderSplits(ctx context.Context, orderID uint) ([]model.OrderSplit, error)

	// Settlements
	CreateSettlement(ctx context.Context, s *model.SettlementBatch) error
	GetSettlement(ctx context.Context, id uint) (*model.SettlementBatch, error)
	ListSettlements(ctx context.Context, merchantID string) ([]model.SettlementBatch, error)
	UpdateSettlement(ctx context.Context, s *model.SettlementBatch) error
	CreateSettlementOrder(ctx context.Context, s *model.SettlementOrder) error
	ListSettlementOrders(ctx context.Context, batchID uint) ([]model.SettlementOrder, error)

	// Refunds
	CreateRefund(ctx context.Context, r *model.Refund) error
	GetRefund(ctx context.Context, id uint) (*model.Refund, error)
	GetRefundByOutNo(ctx context.Context, outNo string) (*model.Refund, error)
	UpdateRefund(ctx context.Context, r *model.Refund) error
	ListRefunds(ctx context.Context, orderID uint) ([]model.Refund, error)

	// Split Records
	CreateSplitRecord(ctx context.Context, s *model.SplitRecord) error
	GetSplitRecord(ctx context.Context, id uint) (*model.SplitRecord, error)
	UpdateSplitRecord(ctx context.Context, s *model.SplitRecord) error
	ListSplitRecords(ctx context.Context, batchID uint) ([]model.SplitRecord, error)

	// Reconciliation
	CreateReconciliation(ctx context.Context, r *model.Reconciliation) error
	GetReconciliation(ctx context.Context, id uint) (*model.Reconciliation, error)
	ListReconciliations(ctx context.Context, channel string, date time.Time) ([]model.Reconciliation, error)
	UpdateReconciliation(ctx context.Context, r *model.Reconciliation) error
	CreateReconciliationDetail(ctx context.Context, d *model.ReconciliationDetail) error
	ListReconciliationDetails(ctx context.Context, reconID uint) ([]model.ReconciliationDetail, error)
	UpdateReconciliationDetail(ctx context.Context, d *model.ReconciliationDetail) error

	// Crypto
	CreateRedemption(ctx context.Context, r *model.CryptoRedemption) error
	GetRedemption(ctx context.Context, id uint) (*model.CryptoRedemption, error)
	GetRedemptionByTxHash(ctx context.Context, txHash string) (*model.CryptoRedemption, error)
	UpdateRedemption(ctx context.Context, r *model.CryptoRedemption) error
	CreateWallet(ctx context.Context, w *model.CryptoWallet) error
	GetWallet(ctx context.Context, id uint) (*model.CryptoWallet, error)
	GetWalletByAddress(ctx context.Context, address string) (*model.CryptoWallet, error)
	UpdateWallet(ctx context.Context, w *model.CryptoWallet) error
	CreateTransaction(ctx context.Context, t *model.CryptoTransaction) error
	GetTransaction(ctx context.Context, id uint) (*model.CryptoTransaction, error)
	GetTransactionByTxHash(ctx context.Context, txHash string) (*model.CryptoTransaction, error)
	UpdateTransaction(ctx context.Context, t *model.CryptoTransaction) error

	// KYC/AML
	CreateKYC(ctx context.Context, k *model.KYCVerification) error
	GetKYC(ctx context.Context, id uint) (*model.KYCVerification, error)
	GetKYCByUser(ctx context.Context, userID uint) (*model.KYCVerification, error)
	UpdateKYC(ctx context.Context, k *model.KYCVerification) error
	CreateAML(ctx context.Context, a *model.AMLScreening) error
	GetAML(ctx context.Context, id uint) (*model.AMLScreening, error)
	GetAMLByUser(ctx context.Context, userID uint) (*model.AMLScreening, error)
	UpdateAML(ctx context.Context, a *model.AMLScreening) error

	// Audit
	CreateAuditLog(ctx context.Context, a *model.AuditLog) error
	ListAuditLogs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]model.AuditLog, error)

	// Ledger
	CreateAccount(ctx context.Context, a *model.Account) error
	GetAccount(ctx context.Context, id uint) (*model.Account, error)
	GetAccountByType(ctx context.Context, accountType string) (*model.Account, error)
	UpdateAccount(ctx context.Context, a *model.Account) error
	CreateLedgerEntry(ctx context.Context, e *model.LedgerEntry) error
	ListLedgerEntries(ctx context.Context, referenceID string) ([]model.LedgerEntry, error)

	// Merchants
	CreateMerchant(ctx context.Context, m *model.Merchant) error
	GetMerchant(ctx context.Context, apiKey string) (*model.Merchant, error)
	GetMerchantByID(ctx context.Context, id uint) (*model.Merchant, error)
	UpdateMerchant(ctx context.Context, m *model.Merchant) error

	// Provider
	CreateProviderConfig(ctx context.Context, p *model.ProviderConfig) error
	GetProviderConfig(ctx context.Context, provider string) (*model.ProviderConfig, error)
	UpdateProviderConfig(ctx context.Context, p *model.ProviderConfig) error

	// Risk
	CreateRiskRule(ctx context.Context, r *model.RiskRule) error
	GetRiskRule(ctx context.Context, id uint) (*model.RiskRule, error)
	ListRiskRules(ctx context.Context) ([]model.RiskRule, error)
	UpdateRiskRule(ctx context.Context, r *model.RiskRule) error
	CreateRiskEvent(ctx context.Context, e *model.RiskEvent) error
	GetRiskEvent(ctx context.Context, id uint) (*model.RiskEvent, error)
	ListRiskEvents(ctx context.Context, filters map[string]interface{}, limit int) ([]model.RiskEvent, error)
	UpdateRiskEvent(ctx context.Context, e *model.RiskEvent) error

	// Fee
	CreateFeeConfig(ctx context.Context, f *model.FeeConfig) error
	GetFeeConfig(ctx context.Context, channel string) (*model.FeeConfig, error)
	UpdateFeeConfig(ctx context.Context, f *model.FeeConfig) error

	// Webhooks
	CreateWebhookDelivery(ctx context.Context, w *model.WebhookDelivery) error
	ListPendingWebhooks(ctx context.Context, before time.Time, limit int) ([]model.WebhookDelivery, error)
	UpdateWebhookDelivery(ctx context.Context, w *model.WebhookDelivery) error
}

var ErrNotFound = errors.New("not found")