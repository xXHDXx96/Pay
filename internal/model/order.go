package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// OrderStatus represents the lifecycle of a payment order.
type OrderStatus string

const (
	OrderPending  OrderStatus = "PENDING"
	OrderPaid     OrderStatus = "PAID"
	OrderSettled  OrderStatus = "SETTLED"
	OrderFailed   OrderStatus = "FAILED"
	OrderClosed   OrderStatus = "CLOSED"
	OrderRefunded OrderStatus = "REFUNDED"
)

// PayChannel selects the underlying rail. A single QR code can carry
// multiple channels (one code, many payments).
type PayChannel string

const (
	ChannelAlipay   PayChannel = "ALIPAY"
	ChannelWechat   PayChannel = "WECHAT"
	ChannelUnionpay PayChannel = "UNIONPAY"
	ChannelCrypto   PayChannel = "CRYPTO"
	ChannelBankCard PayChannel = "BANK_CARD"
)

// Order is the canonical merchant-facing record.
type Order struct {
	gorm.Model
	MerchantID    string          `json:"merchant_id" gorm:"index"`
	OutOrderNo    string          `json:"out_order_no" gorm:"uniqueIndex"`
	Subject       string          `json:"subject"`
	AmountCNY     decimal.Decimal `json:"amount_cny"`
	Channel       PayChannel      `json:"channel"`
	Status        OrderStatus     `json:"status" gorm:"index"`
	QRCodeURL     string          `json:"qr_code_url,omitempty"`
	PayURL        string          `json:"pay_url,omitempty"`
	TradeNo       string          `json:"trade_no,omitempty" gorm:"index"`
	TransactionID string          `json:"transaction_id,omitempty"`
	PaidAt        *time.Time      `json:"paid_at,omitempty"`
	SettledAt     *time.Time      `json:"settled_at,omitempty"`
	RefundedAt    *time.Time      `json:"refunded_at,omitempty"`
}

// CryptoRedemption is the on-chain leg of a one-click crypto exchange.
type CryptoRedemption struct {
	gorm.Model
	OrderID       uint            `json:"-" gorm:"index"`
	WalletAddress string          `json:"wallet_address"`
	CryptoAsset   string          `json:"crypto_asset"`
	CryptoAmount  decimal.Decimal `json:"crypto_amount"`
	FiatCNY       float64         `json:"fiat_cny"`
	ExchangeRate  float64         `json:"exchange_rate"`
	TxHash        string          `json:"tx_hash,omitempty" gorm:"index"`
	Network       string          `json:"network,omitempty"`
	Status        string          `json:"status"` // PENDING, BROADCASTED, CONFIRMED, FAILED
	CreatedAt     time.Time       `json:"created_at"`
	ConfirmedAt   *time.Time      `json:"confirmed_at,omitempty"`
}

// Refund represents a partial/full refund of a paid order.
type Refund struct {
	gorm.Model
	OrderID          uint            `json:"-" gorm:"index"`
	OutRefundNo      string          `json:"out_refund_no" gorm:"uniqueIndex"`
	AmountCNY        decimal.Decimal `json:"amount_cny"`
	Amount           float64         `json:"amount,omitempty"`
	Reason           string          `json:"reason"`
	ProviderRefundNo string          `json:"provider_refund_no,omitempty"`
	Status           string          `json:"status"`
	Channel          string          `json:"channel"`
	ChannelStr       string          `json:"channel_str,omitempty"`
	OperatorID       uint            `json:"operator_id,omitempty"`
	FailedReason     string          `json:"failed_reason,omitempty"`
	RefundedAt       *time.Time      `json:"refunded_at,omitempty"`
}

// Merchant represents a registered business account.
type Merchant struct {
	gorm.Model
	Name             string          `json:"name"`
	APIKey           string          `json:"-" gorm:"uniqueIndex"`
	WebhookURL       string          `json:"webhook_url,omitempty"`
	WebhookSecret    string          `json:"-" gorm:"default:''"`
	SettlementWallet string          `json:"settlement_wallet,omitempty"`
	CryptoEnabled    bool            `json:"crypto_enabled"`
	Active           bool            `json:"active" gorm:"default:true"`
	SettlementCycle  string          `json:"settlement_cycle,omitempty"`
	FeeRate          decimal.Decimal `json:"fee_rate"`
	KYCStatus        string          `json:"kyc_status,omitempty"`
	AMLStatus        string          `json:"aml_status,omitempty"`
	RiskLevel        string          `json:"risk_level,omitempty"`
}

// OrderSplit represents a single receiver in a split payment.
type OrderSplit struct {
	gorm.Model
	OrderID    uint    `json:"order_id" gorm:"index"`
	ReceiverID string  `json:"receiver_id"`
	AmountCNY  float64 `json:"amount_cny"`
	Status     string  `json:"status"` // PENDING, PAID, FAILED
	TradeNo    string  `json:"trade_no,omitempty"`
}

// SettlementOrder links an order to a settlement batch.
type SettlementOrder struct {
	gorm.Model
	BatchID      uint            `json:"batch_id" gorm:"index"`
	OrderID      uint            `json:"order_id"`
	AmountCNY    decimal.Decimal `json:"amount_cny"`
	FeeCNY       decimal.Decimal `json:"fee_cny"`
	NetAmountCNY decimal.Decimal `json:"net_amount_cny"`
	Status       string          `json:"status"` // PENDING, SETTLED, FAILED
}

// SplitRecord records a completed split disbursement.
type SplitRecord struct {
	gorm.Model
	OrderID      uint            `json:"order_id"`
	BatchID      uint            `json:"batch_id,omitempty"`
	AccountID    uint            `json:"account_id,omitempty"`
	AccountType  string          `json:"account_type"`
	AmountCNY    decimal.Decimal `json:"amount_cny"`
	NetAmountCNY decimal.Decimal `json:"net_amount_cny"`
	Status       string          `json:"status"` // SUCCESS, FAILED, PENDING
	FailedReason string          `json:"failed_reason,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	ProcessedAt  time.Time       `json:"processed_at"`
}

// Reconciliation is a batch reconciliation record.
type Reconciliation struct {
	gorm.Model
	Channel       string     `json:"channel"`
	TradeDate     time.Time  `json:"trade_date"`
	Date          time.Time  `json:"date"`
	Provider      string     `json:"provider"`
	TotalRemote   int        `json:"total_remote"`
	TotalCount    int        `json:"total_count"`
	TotalAmount   float64    `json:"total_amount"`
	MatchedCount  int        `json:"matched_count"`
	MismatchCount int        `json:"mismatch_count"`
	MissingLocal  int        `json:"missing_local"`
	MissingRemote int        `json:"missing_remote"`
	Status        string     `json:"status"` // PENDING, PROCESSED, COMPLETED, MATCHED, MISMATCH
	FileName      string     `json:"file_name"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// ReconciliationDetail is a line item in a reconciliation.
type ReconciliationDetail struct {
	gorm.Model
	ReconID        uint            `json:"recon_id" gorm:"index"`
	LocalOrderNo   string          `json:"local_order_no,omitempty"`
	RemoteOrderNo  string          `json:"remote_order_no,omitempty"`
	LocalAmount    decimal.Decimal `json:"local_amount,omitempty"`
	RemoteAmount   decimal.Decimal `json:"remote_amount,omitempty"`
	AmountCNY      decimal.Decimal `json:"amount_cny,omitempty"`
	LocalStatus    string          `json:"local_status,omitempty"`
	RemoteStatus   string          `json:"remote_status,omitempty"`
	TradeNo        string          `json:"trade_no,omitempty"`
	Status         string          `json:"status"` // MATCHED, UNMATCHED, MISSING_REMOTE, MISSING_LOCAL, MISMATCH
	Discrepancy    string          `json:"discrepancy,omitempty"`
	Resolved       bool            `json:"resolved"`
	ResolvedBy     uint            `json:"resolved_by,omitempty"`
	ResolutionNote string          `json:"resolution_note,omitempty"`
	ResolvedAt     *time.Time      `json:"resolved_at,omitempty"`
}

// CryptoWallet represents a wallet address for a merchant.
type CryptoWallet struct {
	gorm.Model
	MerchantID    uint            `json:"merchant_id" gorm:"index"`
	Address       string          `json:"address" gorm:"uniqueIndex"`
	WalletAddress string          `json:"wallet_address,omitempty"`
	WalletType    string          `json:"wallet_type,omitempty"`
	Network       string          `json:"network"`
	Asset         string          `json:"asset"`
	CryptoAsset   string          `json:"crypto_asset,omitempty"`
	Balance       decimal.Decimal `json:"balance"`
	IsReserve     bool            `json:"is_reserve"`
	Active        bool            `json:"active" gorm:"default:true"`
}

// CryptoTransaction represents a blockchain transaction.
type CryptoTransaction struct {
	gorm.Model
	OrderID     uint       `json:"order_id" gorm:"index"`
	TxHash      string     `json:"tx_hash" gorm:"uniqueIndex"`
	FromAddress string     `json:"from_address"`
	ToAddress   string     `json:"to_address"`
	Amount      float64    `json:"amount"`
	Asset       string     `json:"asset"`
	Network     string     `json:"network"`
	Status      string     `json:"status"` // PENDING, CONFIRMED, FAILED
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
}

// KYCVerification represents a KYC check for a user.
type KYCVerification struct {
	gorm.Model
	UserID         uint       `json:"user_id" gorm:"index"`
	Status         string     `json:"status"` // PENDING, APPROVED, REJECTED
	Provider       string     `json:"provider"`
	ReferenceID    string     `json:"reference_id"`
	VerifiedName   string     `json:"verified_name,omitempty"`
	VerifiedIDCard string     `json:"verified_id_card,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// AMLScreening represents an AML check for a user.
type AMLScreening struct {
	gorm.Model
	UserID    uint            `json:"user_id" gorm:"index"`
	Status    string          `json:"status"` // PASS, FAIL, REVIEW
	Score     decimal.Decimal `json:"score"`
	Provider  string          `json:"provider"`
	CreatedAt time.Time       `json:"created_at"`
	Matches   string          `json:"matches,omitempty"`
	RiskLevel string          `json:"risk_level,omitempty"`
}

// AuditLog records a security/event audit entry.
type AuditLog struct {
	gorm.Model
	UserID        uint      `json:"user_id,omitempty"`
	Username      string    `json:"username,omitempty"`
	Action        string    `json:"action"`
	ResourceType  string    `json:"resource_type"`
	ResourceID    string    `json:"resource_id,omitempty"`
	BeforeData    string    `json:"before_data,omitempty"`
	AfterData     string    `json:"after_data,omitempty"`
	ChangeSummary string    `json:"change_summary,omitempty"`
	Severity      string    `json:"severity"`
	TraceID       string    `json:"trace_id,omitempty"`
	IPAddress     string    `json:"ip_address,omitempty"`
	UserAgent     string    `json:"user_agent,omitempty"`
	Detail        string    `json:"detail,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// Account is a ledger account (double-entry).
type Account struct {
	gorm.Model
	AccountType string          `json:"account_type" gorm:"uniqueIndex"`
	AccountName string          `json:"account_name"`
	BalanceCNY  decimal.Decimal `json:"balance_cny"`
	Balance     float64         `json:"balance"`
	Currency    string          `json:"currency"`
	Active      bool            `json:"active" gorm:"default:true"`
	FrozenCNY   decimal.Decimal `json:"frozen_cny"`
	ReservedCNY decimal.Decimal `json:"reserved_cny"`
}

// LedgerEntry is a double-entry ledger line item.
type LedgerEntry struct {
	gorm.Model
	AccountID     uint            `json:"account_id" gorm:"index"`
	OrderID       uint            `json:"order_id,omitempty" gorm:"index"`
	ReferenceID   string          `json:"reference_id"`
	ReferenceType string          `json:"reference_type,omitempty"`
	AmountCNY     decimal.Decimal `json:"amount_cny"`
	Amount        float64         `json:"amount"`
	Currency      string          `json:"currency"`
	EntryType     string          `json:"entry_type"` // DEBIT, CREDIT
	Direction     string          `json:"direction"`  // DEBIT, CREDIT
	Description   string          `json:"description"`
	BalanceAfter  decimal.Decimal `json:"balance_after"`
	OperatorID    uint            `json:"operator_id,omitempty"`
	TraceID       string          `json:"trace_id,omitempty"`
	Timestamp     time.Time       `json:"timestamp"`
}

// ProviderConfig stores credentials for a payment provider.
type ProviderConfig struct {
	gorm.Model
	Provider string `json:"provider" gorm:"uniqueIndex"`
	Config   string `json:"config"` // JSON blob
	Active   bool   `json:"active" gorm:"default:true"`
}

// RiskRule defines a risk scoring rule.
type RiskRule struct {
	gorm.Model
	Name      string `json:"name"`
	RuleName  string `json:"rule_name"`
	RuleType  string `json:"rule_type"` // FREQUENCY, AMOUNT, BLACKLIST, DEVICE, GEO
	Condition string `json:"condition"`
	Score     int    `json:"score"`
	Action    string `json:"action"` // BLOCK, REVIEW, ALLOW
	Enabled   bool   `json:"enabled" gorm:"default:true"`
	Active    bool   `json:"active" gorm:"default:true"`
}

// RiskEvent records a triggered risk event.
type RiskEvent struct {
	gorm.Model
	OrderID           uint      `json:"order_id,omitempty" gorm:"index"`
	UserID            uint      `json:"user_id,omitempty"`
	RuleID            uint      `json:"rule_id"`
	RuleName          string    `json:"rule_name,omitempty"`
	Score             int       `json:"score"`
	Action            string    `json:"action"` // BLOCK, REVIEW
	Reason            string    `json:"reason,omitempty"`
	RiskLevel         string    `json:"risk_level"`
	IP                string    `json:"ip,omitempty"`
	DeviceFingerprint string    `json:"device_fingerprint,omitempty"`
	GeoCountry        string    `json:"geo_country,omitempty"`
	GeoCity           string    `json:"geo_city,omitempty"`
	IsVPN             bool      `json:"is_vpn"`
	IsProxy           bool      `json:"is_proxy"`
	IsTor             bool      `json:"is_tor"`
	Timestamp         time.Time `json:"timestamp"`
}

// FeeConfig defines fee configuration per channel.
type FeeConfig struct {
	gorm.Model
	Channel  string  `json:"channel" gorm:"uniqueIndex"`
	Rate     float64 `json:"rate"`
	FixedFee float64 `json:"fixed_fee"`
	Active   bool    `json:"active" gorm:"default:true"`
}

// AccountBalance is a snapshot of account balance.
type AccountBalance struct {
	gorm.Model
	AccountID uint      `json:"account_id" gorm:"index"`
	Balance   float64   `json:"balance"`
	Locked    float64   `json:"locked"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SettlementBatch groups orders settled to a merchant's payout account.
type SettlementBatch struct {
	gorm.Model
	MerchantID      uint            `json:"merchant_id" gorm:"index"`
	MerchantIDStr   string          `json:"merchant_id_str,omitempty"`
	BatchNo         string          `json:"batch_no" gorm:"uniqueIndex"`
	TotalAmountCNY  decimal.Decimal `json:"total_amount_cny"`
	TotalAmount     float64         `json:"total_amount,omitempty"`
	OrderCount      int             `json:"order_count"`
	FeeCNY          decimal.Decimal `json:"fee_cny"`
	NetAmountCNY    decimal.Decimal `json:"net_amount_cny"`
	Status          string          `json:"status"` // PENDING, PROCESSING, PAID, FAILED
	PaidAt          *time.Time      `json:"paid_at,omitempty"`
	BankCode        string          `json:"bank_code,omitempty"`
	BankAccountName string          `json:"bank_account_name,omitempty"`
	BankAccountNo   string          `json:"bank_account_no,omitempty"`
	BankSeqNo       string          `json:"bank_seq_no,omitempty"`
	FailedReason    string          `json:"failed_reason,omitempty"`
	RetryCount      int             `json:"retry_count"`
	ConfirmedBy     uint            `json:"confirmed_by,omitempty"`
	ConfirmedAt     *time.Time      `json:"confirmed_at,omitempty"`
}

// WebhookDelivery records an outbound webhook attempt (retry-able).
type WebhookDelivery struct {
	gorm.Model
	MerchantID  string     `json:"merchant_id" gorm:"index"`
	EventType   string     `json:"event_type"`
	Payload     string     `json:"payload"`
	TargetURL   string     `json:"target_url"`
	Attempts    int        `json:"attempts"`
	NextRetryAt *time.Time `json:"next_retry_at,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	Status      string     `json:"status"` // PENDING, DELIVERED, DEAD
}
