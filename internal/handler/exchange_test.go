package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"alipay-payment/internal/exchange"
	"alipay-payment/internal/model"
	"alipay-payment/internal/store"
	"github.com/shopspring/decimal"
)

func TestExchangeQuote(t *testing.T) {
	s := store.NewMemory()
	redem := exchange.NewRedemptions(exchange.NewOracle())
	h := NewExchangeHandler(s, redem)

	body := `{"asset":"USDT","fiat_cny":71.5,"order_id":"1"}`
	req := httptest.NewRequest("POST", "/exchange/quote", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["crypto_asset"] != "USDT" {
		t.Fatalf("expected USDT, got %v", resp["crypto_asset"])
	}
	// crypto_amount should be 10 (71.5 / 7.15 = 10)
	amtStr, ok := resp["crypto_amount"].(string)
	if !ok {
		t.Fatalf("expected string for crypto_amount, got %T: %v", resp["crypto_amount"], resp["crypto_amount"])
	}
	if amtStr != "10" {
		t.Fatalf("expected 10 USDT, got %s", amtStr)
	}
}

func TestExchangeRedeem(t *testing.T) {
	s := store.NewMemory()
	redem := exchange.NewRedemptions(exchange.NewOracle())
	redem.SetBroadcaster(&mockBroadcaster{})
	h := NewExchangeHandler(s, redem)

	seed := &model.CryptoRedemption{OrderID: 1, CryptoAsset: "USDT", CryptoAmount: decimal.NewFromInt(10), FiatCNY: 71.5, ExchangeRate: 7.15, Status: "PENDING"}
	if err := s.CreateRedemption(context.Background(), seed); err != nil {
		t.Fatalf("seed: %v", err)
	}
	got, err := s.GetRedemption(context.Background(), seed.ID)
	t.Logf("seed GetRedemption(%d): err=%v, got=%+v", seed.ID, err, got)

	body := `{"order_id":"` + fmt.Sprintf("%d", seed.ID) + `","wallet_addr":"0xdest"}`
	req := httptest.NewRequest("POST", "/exchange/redeem", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	t.Logf("response code: %d, body: %s", w.Code, w.Body.String())

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "CONFIRMED" {
		t.Fatalf("expected CONFIRMED, got %v", resp["status"])
	}
	if resp["tx_hash"] != "0xstub" {
		t.Fatalf("expected tx hash, got %v", resp["tx_hash"])
	}
}

type mockBroadcaster struct{}

func (m *mockBroadcaster) Transfer(ctx context.Context, toAddress string, amount *big.Int) (string, error) {
	return "0xstub", nil
}
