package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"alipay-payment/internal/model"
	"alipay-payment/internal/store"
	"github.com/shopspring/decimal"
)

type PayRequest struct {
	MerchantID string           `json:"merchant_id"`
	OutOrderNo string           `json:"out_order_no"`
	Subject    string           `json:"subject"`
	AmountCNY  float64          `json:"amount_cny"`
	Channel    model.PayChannel `json:"channel"`
	WalletAddr string           `json:"wallet_addr,omitempty"`
}

type PayHandler struct {
	store      store.Store
	alipayQR   func(outOrderNo, subject string, amountCNY float64) (string, error)
	wechatQR   func(outOrderNo, subject string, amountCNY float64) (string, error)
	bankcardQR func(outOrderNo, subject string, amountCNY float64) (string, error)
	unionPayQR func(outOrderNo, subject string, amountCNY float64) (string, error)
}

func NewPayHandler(s store.Store) *PayHandler {
	return &PayHandler{store: s}
}

func (h *PayHandler) WithAlipayQR(fn func(outOrderNo, subject string, amountCNY float64) (string, error)) *PayHandler {
	h.alipayQR = fn
	return h
}

func (h *PayHandler) WithWechatQR(fn func(outOrderNo, subject string, amountCNY float64) (string, error)) *PayHandler {
	h.wechatQR = fn
	return h
}

func (h *PayHandler) WithBankcardQR(fn func(outOrderNo, subject string, amountCNY float64) (string, error)) *PayHandler {
	h.bankcardQR = fn
	return h
}

func (h *PayHandler) WithUnionPayQR(fn func(outOrderNo, subject string, amountCNY float64) (string, error)) *PayHandler {
	h.unionPayQR = fn
	return h
}

func (h *PayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req PayRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.OutOrderNo == "" || req.Subject == "" || req.AmountCNY <= 0 {
		http.Error(w, "out_order_no, subject and amount_cny>0 are required", http.StatusBadRequest)
		return
	}
	if req.Channel == "" {
		req.Channel = model.ChannelAlipay
	}

	if existing, err := h.store.GetOrderByOutNo(r.Context(), req.OutOrderNo); err == nil {
		encodeJSON(w, existing)
		return
	}

	order := &model.Order{
		MerchantID: req.MerchantID,
		OutOrderNo: req.OutOrderNo,
		Subject:    req.Subject,
		AmountCNY:  decimal.NewFromFloat(req.AmountCNY),
		Channel:    req.Channel,
		Status:     model.OrderPending,
	}

	switch req.Channel {
	case model.ChannelCrypto:
		order.QRCodeURL = "crypto://pay/" + req.WalletAddr
	case model.ChannelAlipay:
		if h.alipayQR != nil {
			qr, err := h.alipayQR(req.OutOrderNo, req.Subject, req.AmountCNY)
			if err != nil {
				http.Error(w, "alipay precreate failed: "+err.Error(), http.StatusBadGateway)
				return
			}
			order.QRCodeURL = qr
		}
	case model.ChannelWechat:
		if h.wechatQR != nil {
			qr, err := h.wechatQR(req.OutOrderNo, req.Subject, req.AmountCNY)
			if err != nil {
				http.Error(w, "wechat precreate failed: "+err.Error(), http.StatusBadGateway)
				return
			}
			order.QRCodeURL = qr
		}
	case model.ChannelBankCard:
		if h.unionPayQR != nil {
			qr, err := h.unionPayQR(req.OutOrderNo, req.Subject, req.AmountCNY)
			if err != nil {
				http.Error(w, "unionpay precreate failed: "+err.Error(), http.StatusBadGateway)
				return
			}
			order.QRCodeURL = qr
		} else if h.bankcardQR != nil {
			qr, err := h.bankcardQR(req.OutOrderNo, req.Subject, req.AmountCNY)
			if err != nil {
				http.Error(w, "bankcard precreate failed: "+err.Error(), http.StatusBadGateway)
				return
			}
			order.QRCodeURL = qr
		}
	default:
		order.QRCodeURL = string(req.Channel) + "://pay/" + req.OutOrderNo
	}

	if err := h.store.CreateOrder(r.Context(), order); err != nil {
		http.Error(w, "create order: "+err.Error(), http.StatusInternalServerError)
		return
	}
	encodeJSON(w, order)
}

type OrdersHandler struct{ store store.Store }

func NewOrdersHandler(s store.Store) *OrdersHandler { return &OrdersHandler{store: s} }

func (h *OrdersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	orders, err := h.store.ListOrders(r.Context(), r.URL.Query().Get("merchant_id"))
	if err != nil {
		http.Error(w, "list orders: "+err.Error(), http.StatusInternalServerError)
		return
	}
	encodeJSON(w, map[string]any{"orders": orders, "count": len(orders)})
}

type OrderHandler struct{ store store.Store }

func NewOrderHandler(s store.Store) *OrderHandler { return &OrderHandler{store: s} }

func (h *OrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/orders/")
	if idStr == "" || strings.Contains(idStr, "/") {
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}
	o, err := h.store.GetOrder(r.Context(), uint(id))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	encodeJSON(w, o)
}

func AmountToFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
