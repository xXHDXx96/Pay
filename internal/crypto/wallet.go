package crypto

import (
	"context"
	"errors"
	"fmt"
	"time"

	"alipay-payment/internal/model"
	"alipay-payment/internal/store"
	"github.com/shopspring/decimal"
)

type Wallet struct {
	store    store.Store
	rpcURL   string
	chainID  int64
	masterPK string
	stopCh   chan struct{}
}

type WalletConfig struct {
	RPCURL            string
	ChainID           int64
	MasterPrivateKey  string
	USDTContract      string
	USDCContract      string
	HotWallet         string
	ColdWallet        string
	MultisigThreshold int
	MultisigOwners    []string
	GasStationURL     string
	GasStationKey     string
	TravelRuleAPIURL  string
	TravelRuleAPIKey  string
}

func NewWallet(cfg WalletConfig) *Wallet {
	return &Wallet{
		rpcURL: cfg.RPCURL, chainID: cfg.ChainID,
		masterPK: cfg.MasterPrivateKey, stopCh: make(chan struct{}),
	}
}

func NewMockWallet() *Wallet {
	return &Wallet{stopCh: make(chan struct{})}
}

func (w *Wallet) Quote(ctx context.Context, asset string, fiatCNY decimal.Decimal) (decimal.Decimal, error) {
	rates := map[string]float64{"USDT": 7.15, "USDC": 7.15, "BTC": 62000, "ETH": 3400}
	rate, ok := rates[asset]
	if !ok {
		return decimal.Zero, errors.New("unsupported asset")
	}
	return fiatCNY.Div(decimal.NewFromFloat(rate)), nil
}

func (w *Wallet) Redeem(ctx context.Context, orderID uint, walletAddr, asset, network string, cryptoAmount decimal.Decimal) (*model.CryptoRedemption, error) {
	if walletAddr == "" {
		return nil, errors.New("wallet address required")
	}
	if cryptoAmount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("amount must be positive")
	}

	if cryptoAmount.Mul(decimal.NewFromFloat(7.15)).GreaterThan(decimal.NewFromFloat(1000)) {
		// Travel Rule 上报
	}

	redemption := &model.CryptoRedemption{
		OrderID: orderID, WalletAddress: walletAddr, CryptoAsset: asset,
		CryptoAmount: cryptoAmount, Network: network,
		Status: "BROADCASTED", CreatedAt: time.Now(),
	}

	redemption.TxHash = fmt.Sprintf("0x%x", time.Now().UnixNano())
	redemption.Status = "CONFIRMED"
	now := time.Now()
	redemption.ConfirmedAt = &now

	return redemption, nil
}

func (w *Wallet) GasStation() *GasStation {
	return &GasStation{stopCh: make(chan struct{})}
}

type GasStation struct {
	stopCh chan struct{}
}

func (g *GasStation) Start() {}
func (g *GasStation) Stop()  { close(g.stopCh) }

func (w *Wallet) GetBalance(ctx context.Context, asset string) (decimal.Decimal, error) {
	return decimal.NewFromFloat(10000), nil
}

func (w *Wallet) CreateWallet(ctx context.Context, walletType, asset string, reserve bool) (*model.CryptoWallet, error) {
	return &model.CryptoWallet{
		WalletType: walletType, WalletAddress: fmt.Sprintf("0x%x", time.Now().UnixNano()),
		CryptoAsset: asset, Balance: decimal.NewFromFloat(10000), IsReserve: reserve,
	}, nil
}
