package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"alipay-payment/internal/model"
	"alipay-payment/internal/store"
	"github.com/shopspring/decimal"
)

type QRPaymentRequest struct {
	MerchantID    string  `json:"merchant_id"`
	OutOrderNo    string  `json:"out_order_no"`
	Subject       string  `json:"subject"`
	AmountCNY     float64 `json:"amount_cny"`
	Channel       string  `json:"channel"`
	WalletAddr    string  `json:"wallet_addr,omitempty"`
	ExpireMinutes int     `json:"expire_minutes,omitempty"`
	CallbackURL   string  `json:"callback_url,omitempty"`
}

type QRPaymentResponse struct {
	OrderID      uint      `json:"order_id"`
	OutOrderNo   string    `json:"out_order_no"`
	QRCodeURL    string    `json:"qr_code_url"`
	QRCodeBase64 string    `json:"qr_code_base64,omitempty"`
	Channel      string    `json:"channel"`
	AmountCNY    float64   `json:"amount_cny"`
	Status       string    `json:"status"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type PaymentStatusResponse struct {
	OrderID       uint       `json:"order_id"`
	OutOrderNo    string     `json:"out_order_no"`
	Status        string     `json:"status"`
	AmountCNY     float64    `json:"amount_cny"`
	Channel       string     `json:"channel"`
	CreatedAt     time.Time  `json:"created_at"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	TradeNo       string     `json:"trade_no,omitempty"`
	TransactionID string     `json:"transaction_id,omitempty"`
	QRCodeURL     string     `json:"qr_code_url,omitempty"`
	ExpiresAt     time.Time  `json:"expires_at"`
}

type QRCodeHandler struct {
	store   store.Store
	qrFuncs map[model.PayChannel]func(outOrderNo, subject string, amountCNY float64) (string, error)
	mu      sync.RWMutex
	orders  map[string]*model.Order
}

func NewQRCodeHandler(s store.Store) *QRCodeHandler {
	return &QRCodeHandler{
		store:   s,
		qrFuncs: make(map[model.PayChannel]func(outOrderNo, subject string, amountCNY float64) (string, error)),
		orders:  make(map[string]*model.Order),
	}
}

func (h *QRCodeHandler) WithChannel(ch model.PayChannel, fn func(outOrderNo, subject string, amountCNY float64) (string, error)) *QRCodeHandler {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.qrFuncs[ch] = fn
	return h
}

func (h *QRCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasPrefix(path, "/aggregation") {
		if r.Method == http.MethodPost {
			h.CreateAggregationPayment(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	if strings.HasPrefix(path, "/resolve") {
		if r.Method == http.MethodGet {
			h.ResolveAggregation(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	switch r.Method {
	case http.MethodPost:
		h.CreatePayment(w, r)
	case http.MethodGet:
		h.QueryPayment(w, r)
	case http.MethodPut:
		h.CancelPayment(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *QRCodeHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req QRPaymentRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.OutOrderNo == "" || req.Subject == "" || req.AmountCNY <= 0 {
		http.Error(w, "out_order_no, subject and amount_cny>0 required", http.StatusBadRequest)
		return
	}
	if req.Channel == "" {
		req.Channel = string(model.ChannelAlipay)
	}
	if req.ExpireMinutes == 0 {
		req.ExpireMinutes = 15
	}

	channel := model.PayChannel(req.Channel)
	h.mu.RLock()
	qrFunc, ok := h.qrFuncs[channel]
	h.mu.RUnlock()
	if !ok {
		qrFunc, ok = h.qrFuncs[model.ChannelAlipay]
		if !ok {
			http.Error(w, "channel not supported", http.StatusBadRequest)
			return
		}
		channel = model.ChannelAlipay
	}

	expiredAt := time.Now().Add(time.Duration(req.ExpireMinutes) * time.Minute)

	qrURL, err := qrFunc(req.OutOrderNo, req.Subject, req.AmountCNY)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s precreate failed: %v", channel, err.Error()), http.StatusBadGateway)
		return
	}

	b64QR, _ := generateTinyQR(qrURL)

	order := &model.Order{
		MerchantID: req.MerchantID,
		OutOrderNo: req.OutOrderNo,
		Subject:    req.Subject,
		AmountCNY:  decimal.NewFromFloat(req.AmountCNY),
		Channel:    channel,
		Status:     model.OrderPending,
		QRCodeURL:  qrURL,
		PaidAt:     nil,
		SettledAt:  nil,
	}

	if err := h.store.CreateOrder(r.Context(), order); err != nil {
		http.Error(w, "create order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.mu.Lock()
	h.orders[req.OutOrderNo] = order
	h.mu.Unlock()

	encodeJSON(w, QRPaymentResponse{
		OrderID:      order.ID,
		OutOrderNo:   order.OutOrderNo,
		QRCodeURL:    qrURL,
		QRCodeBase64: b64QR,
		Channel:      string(channel),
		AmountCNY:    req.AmountCNY,
		Status:       string(order.Status),
		ExpiresAt:    expiredAt,
		CreatedAt:    time.Now(),
	})
}

func (h *QRCodeHandler) QueryPayment(w http.ResponseWriter, r *http.Request) {
	outOrderNo := r.URL.Query().Get("out_order_no")
	if outOrderNo == "" {
		http.Error(w, "out_order_no required", http.StatusBadRequest)
		return
	}

	order, err := h.store.GetOrderByOutNo(r.Context(), outOrderNo)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	var paidAt *time.Time
	if order.PaidAt != nil {
		paidAt = order.PaidAt
	}

	expire := order.CreatedAt.Add(15 * time.Minute)

	encodeJSON(w, PaymentStatusResponse{
		OrderID:       order.ID,
		OutOrderNo:    order.OutOrderNo,
		Status:        string(order.Status),
		AmountCNY:     order.AmountCNY.InexactFloat64(),
		Channel:       string(order.Channel),
		CreatedAt:     order.CreatedAt,
		PaidAt:        paidAt,
		TradeNo:       order.TradeNo,
		TransactionID: order.TransactionID,
		QRCodeURL:     order.QRCodeURL,
		ExpiresAt:     expire,
	})
}

func (h *QRCodeHandler) CancelPayment(w http.ResponseWriter, r *http.Request) {
	outOrderNo := r.URL.Query().Get("out_order_no")
	if outOrderNo == "" {
		http.Error(w, "out_order_no required", http.StatusBadRequest)
		return
	}

	order, err := h.store.GetOrderByOutNo(r.Context(), outOrderNo)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	if order.Status != model.OrderPending {
		http.Error(w, "order cannot be cancelled in current status", http.StatusBadRequest)
		return
	}

	var reason string
	if err := decodeJSON(r, &struct {
		Reason string `json:"reason"`
	}{Reason: ""}); err == nil {
		// parse reason if body provided
	}
	_ = reason

	order.Status = model.OrderClosed
	if err := h.store.UpdateOrder(r.Context(), order); err != nil {
		http.Error(w, "update order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	encodeJSON(w, map[string]any{
		"order_id":     order.ID,
		"out_order_no": order.OutOrderNo,
		"status":       string(order.Status),
	})
}

type AggregationPaymentRequest struct {
	MerchantID    string  `json:"merchant_id"`
	OutOrderNo    string  `json:"out_order_no"`
	Subject       string  `json:"subject"`
	AmountCNY     float64 `json:"amount_cny"`
	ExpireMinutes int     `json:"expire_minutes,omitempty"`
	CallbackURL   string  `json:"callback_url,omitempty"`
}

type AggregationPaymentResponse struct {
	OrderID    uint          `json:"order_id"`
	OutOrderNo string        `json:"out_order_no"`
	QRCodeURL  string        `json:"qr_code_url"`
	Channels   []ChannelInfo `json:"channels"`
	Status     string        `json:"status"`
	ExpiresAt  time.Time     `json:"expires_at"`
	CreatedAt  time.Time     `json:"created_at"`
}

type ChannelInfo struct {
	Channel   string `json:"channel"`
	QRCodeURL string `json:"qr_code_url,omitempty"`
}

func (h *QRCodeHandler) CreateAggregationPayment(w http.ResponseWriter, r *http.Request) {
	var req AggregationPaymentRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.OutOrderNo == "" || req.Subject == "" || req.AmountCNY <= 0 {
		http.Error(w, "out_order_no, subject and amount_cny>0 required", http.StatusBadRequest)
		return
	}
	if req.ExpireMinutes == 0 {
		req.ExpireMinutes = 15
	}

	expiredAt := time.Now().Add(time.Duration(req.ExpireMinutes) * time.Minute)

	h.mu.RLock()
	qrFuncs := make(map[model.PayChannel]func(string, string, float64) (string, error))
	for ch, fn := range h.qrFuncs {
		qrFuncs[ch] = fn
	}
	h.mu.RUnlock()

	if len(qrFuncs) == 0 {
		http.Error(w, "no payment channels configured", http.StatusServiceUnavailable)
		return
	}

	order := &model.Order{
		MerchantID: req.MerchantID,
		OutOrderNo: req.OutOrderNo,
		Subject:    req.Subject,
		AmountCNY:  decimal.NewFromFloat(req.AmountCNY),
		Status:     model.OrderPending,
	}

	if err := h.store.CreateOrder(r.Context(), order); err != nil {
		http.Error(w, "create order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	channels := make([]ChannelInfo, 0, len(qrFuncs))
	for ch, fn := range qrFuncs {
		qrURL, err := fn(req.OutOrderNo, req.Subject, req.AmountCNY)
		if err != nil {
			continue
		}
		channels = append(channels, ChannelInfo{
			Channel:   string(ch),
			QRCodeURL: qrURL,
		})
	}

	if len(channels) == 0 {
		order.Status = model.OrderFailed
		_ = h.store.UpdateOrder(r.Context(), order)
		http.Error(w, "failed to generate QR codes for any channel", http.StatusServiceUnavailable)
		return
	}

	primaryChannel := channels[0]
	b64QR, _ := generateTinyQR(primaryChannel.QRCodeURL)

	order.Channel = model.PayChannel(primaryChannel.Channel)
	order.QRCodeURL = b64QR
	order.SettledAt = nil
	if err := h.store.UpdateOrder(r.Context(), order); err != nil {
		http.Error(w, "update order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	encodeJSON(w, AggregationPaymentResponse{
		OrderID:    order.ID,
		OutOrderNo: order.OutOrderNo,
		QRCodeURL:  b64QR,
		Channels:   channels,
		Status:     string(order.Status),
		ExpiresAt:  expiredAt,
		CreatedAt:  time.Now(),
	})
}

func (h *QRCodeHandler) ResolveAggregation(w http.ResponseWriter, r *http.Request) {
	outOrderNo := r.URL.Query().Get("out_order_no")
	if outOrderNo == "" {
		http.Error(w, "out_order_no required", http.StatusBadRequest)
		return
	}
	channel := r.URL.Query().Get("channel")

	order, err := h.store.GetOrderByOutNo(r.Context(), outOrderNo)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	h.mu.RLock()
	qrFuncs := make(map[model.PayChannel]func(string, string, float64) (string, error))
	for ch, fn := range h.qrFuncs {
		qrFuncs[ch] = fn
	}
	h.mu.RUnlock()

	if channel != "" {
		ch := model.PayChannel(channel)
		if fn, ok := qrFuncs[ch]; ok {
			qrURL, err := fn(outOrderNo, order.Subject, order.AmountCNY.InexactFloat64())
			if err != nil {
				http.Error(w, "channel precreate failed: "+err.Error(), http.StatusBadGateway)
				return
			}
			encodeJSON(w, map[string]any{
				"order_id":     order.ID,
				"out_order_no": order.OutOrderNo,
				"qr_code_url":  qrURL,
				"status":       string(order.Status),
			})
			return
		}
	}

	channels := make([]ChannelInfo, 0, len(qrFuncs))
	for ch := range qrFuncs {
		channels = append(channels, ChannelInfo{Channel: string(ch)})
	}
	encodeJSON(w, map[string]any{
		"order_id":     order.ID,
		"out_order_no": order.OutOrderNo,
		"status":       string(order.Status),
		"channels":     channels,
	})
}

type WebhookHandler struct {
	store  store.Store
	secret string
}

type WebhookEvent struct {
	OrderID    uint      `json:"order_id"`
	OutOrderNo string    `json:"out_order_no"`
	Status     string    `json:"status"`
	AmountCNY  float64   `json:"amount_cny"`
	TradeNo    string    `json:"trade_no,omitempty"`
	PaidAt     time.Time `json:"paid_at"`
	Timestamp  int64     `json:"timestamp"`
	Nonce      string    `json:"nonce"`
}

func NewWebhookHandler(s store.Store, secret string) *WebhookHandler {
	return &WebhookHandler{store: s, secret: secret}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var event WebhookEvent
	if err := decodeJSON(r, &event); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if h.secret != "" {
		sig := r.Header.Get("X-Signature")
		if sig == "" {
			http.Error(w, "signature missing", http.StatusUnauthorized)
			return
		}
		expected := computeHmacSHA256(sig, h.secret)
		if !hmacEqual(expected, sig) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
	}

	order, err := h.store.GetOrderByOutNo(r.Context(), event.OutOrderNo)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	if event.Status == "PAID" || event.Status == "SUCCESS" {
		order.Status = model.OrderPaid
		order.TradeNo = event.TradeNo
		paidAt := time.Now()
		order.PaidAt = &paidAt
	} else if event.Status == "FAILED" {
		order.Status = model.OrderFailed
	}

	if err := h.store.UpdateOrder(r.Context(), order); err != nil {
		http.Error(w, "update order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	encodeJSON(w, map[string]string{"status": "ok"})
}

func generateTinyQR(data string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString(b)
	return fmt.Sprintf("https://api.qrserver.com/v1/api/?size=200x200&data=%s&token=%s", data, encoded), nil
}

func computeHmacSHA256(data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func hmacEqual(a, b string) bool {
	return hmac.Equal([]byte(a), []byte(b))
}
