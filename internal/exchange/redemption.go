package exchange

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"alipay-payment/internal/model"
	"github.com/shopspring/decimal"
)

type RateSource interface {
	Price(ctx context.Context, asset string) (float64, error)
}

type Oracle struct {
	mu    sync.RWMutex
	rates map[string]float64
}

func NewOracle() *Oracle {
	return &Oracle{rates: map[string]float64{"USDT": 7.15, "BTC": 62000, "ETH": 3400}}
}

func (o *Oracle) Set(asset string, price float64) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.rates[asset] = price
}

func (o *Oracle) Price(_ context.Context, asset string) (float64, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	p, ok := o.rates[asset]
	if !ok {
		return 0, fmt.Errorf("unknown asset %s", asset)
	}
	return p, nil
}

type Broadcaster interface {
	Transfer(ctx context.Context, toAddress string, amount *big.Int) (string, error)
}

type Redemptions struct {
	rates       RateSource
	broadcaster Broadcaster
}

func NewRedemptions(rates RateSource) *Redemptions {
	return &Redemptions{
		rates:       rates,
		broadcaster: &noopBroadcaster{},
	}
}

func (r *Redemptions) SetBroadcaster(b Broadcaster) { r.broadcaster = b }

type noopBroadcaster struct{}

func (n *noopBroadcaster) Transfer(ctx context.Context, toAddress string, amount *big.Int) (string, error) {
	return "", errors.New("broadcaster not configured")
}

func (r *Redemptions) Quote(ctx context.Context, asset string, fiatCNY float64) (model.CryptoRedemption, error) {
	if fiatCNY <= 0 {
		return model.CryptoRedemption{}, errors.New("fiat amount must be positive")
	}
	price, err := r.rates.Price(ctx, asset)
	if err != nil {
		return model.CryptoRedemption{}, err
	}
	return model.CryptoRedemption{
		CryptoAsset:  asset,
		CryptoAmount: decimal.NewFromFloat(fiatCNY / price),
		FiatCNY:      fiatCNY,
		ExchangeRate: price,
		Status:       "PENDING",
		CreatedAt:    time.Now(),
	}, nil
}

func (r *Redemptions) Redeem(ctx context.Context, red *model.CryptoRedemption, toWallet, network string) (model.CryptoRedemption, error) {
	if toWallet == "" {
		return model.CryptoRedemption{}, errors.New("wallet address required")
	}
	if red.CryptoAmount.LessThanOrEqual(decimal.Zero) {
		return model.CryptoRedemption{}, errors.New("crypto amount must be positive")
	}
	if network == "" {
		network = defaultNetwork(red.CryptoAsset)
	}
	red.WalletAddress = toWallet
	red.Network = network
	red.Status = "BROADCASTED"

	amountWei := big.NewInt(int64(red.CryptoAmount.InexactFloat64() * 1e18)) // assume 18 decimals for USDT
	txHash, err := r.broadcaster.Transfer(ctx, toWallet, amountWei)
	if err != nil {
		red.Status = "FAILED"
		return *red, fmt.Errorf("broadcast: %w", err)
	}
	red.TxHash = txHash
	red.Status = "CONFIRMED"
	now := time.Now()
	red.ConfirmedAt = &now
	return *red, nil
}

func defaultNetwork(asset string) string {
	switch asset {
	case "BTC":
		return "BTC"
	case "ETH", "USDT":
		return "ERC20"
	default:
		return "ERC20"
	}
}
