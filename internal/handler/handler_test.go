package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"alipay-payment/internal/store"
)

func TestPayHandler_CryptoQR(t *testing.T) {
	s := store.NewMemory()
	h := NewPayHandler(s)

	body := `{"merchant_id":"m1","out_order_no":"ord-1","subject":"Test","amount_cny":50,"channel":"CRYPTO","wallet_addr":"0xabc"}`
	req := httptest.NewRequest("POST", "/pay", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["channel"] != "CRYPTO" {
		t.Fatalf("expected CRYPTO channel, got %v", resp["channel"])
	}
	if resp["status"] != "PENDING" {
		t.Fatalf("expected PENDING, got %v", resp["status"])
	}
	if resp["qr_code_url"] != "crypto://pay/0xabc" {
		t.Fatalf("unexpected qr: %v", resp["qr_code_url"])
	}
}

func TestPayHandler_Idempotent(t *testing.T) {
	s := store.NewMemory()
	h := NewPayHandler(s)
	body := `{"merchant_id":"m1","out_order_no":"dup","subject":"X","amount_cny":10,"channel":"WECHAT"}`

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/pay", strings.NewReader(body))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("iter %d: expected 200 got %d", i, w.Code)
		}
	}
	orders, _ := s.ListOrders(context.Background(), "m1")
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
}

func TestOrderHandler_InvalidID(t *testing.T) {
	s := store.NewMemory()
	h := NewOrderHandler(s)
	req := httptest.NewRequest("GET", "/orders/nope", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid ID, got %d", w.Code)
	}
}

func TestOrderHandler_NotFound(t *testing.T) {
	s := store.NewMemory()
	h := NewOrderHandler(s)
	req := httptest.NewRequest("GET", "/orders/999", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
