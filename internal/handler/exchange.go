package handler

import (
	"net/http"
	"strconv"

	"alipay-payment/internal/exchange"
	"alipay-payment/internal/store"
)

type ExchangeRequest struct {
	OrderID    string `json:"order_id"`
	WalletAddr string `json:"wallet_addr"`
	Asset      string `json:"asset"`
	Network    string `json:"network,omitempty"`
}

type ExchangeHandler struct {
	store store.Store
	redem *exchange.Redemptions
}

func NewExchangeHandler(s store.Store, redem *exchange.Redemptions) *ExchangeHandler {
	return &ExchangeHandler{store: s, redem: redem}
}

func (h *ExchangeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/exchange/quote":
		h.Quote(w, r)
	case "/exchange/redeem":
		h.Redeem(w, r)
	case "/exchange":
		h.GetRedemption(w, r)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (h *ExchangeHandler) Quote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Asset   string  `json:"asset"`
		FiatCNY float64 `json:"fiat_cny"`
		OrderID string  `json:"order_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Asset == "" || req.FiatCNY <= 0 {
		http.Error(w, "asset and fiat_cny>0 required", http.StatusBadRequest)
		return
	}
	red, err := h.redem.Quote(r.Context(), req.Asset, req.FiatCNY)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.OrderID != "" {
		if oid, err := strconv.ParseUint(req.OrderID, 10, 64); err == nil {
			red.OrderID = uint(oid)
		}
	}
	if err := h.store.CreateRedemption(r.Context(), &red); err != nil {
		http.Error(w, "store: "+err.Error(), http.StatusInternalServerError)
		return
	}
	encodeJSON(w, red)
}

func (h *ExchangeHandler) Redeem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ExchangeRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	oid, err := strconv.ParseUint(req.OrderID, 10, 64)
	if err != nil {
		http.Error(w, "invalid order_id", http.StatusBadRequest)
		return
	}
	if req.WalletAddr == "" {
		http.Error(w, "wallet_addr required", http.StatusBadRequest)
		return
	}
	red, err := h.store.GetRedemption(r.Context(), uint(oid))
	if err != nil {
		http.Error(w, "redemption not found", http.StatusNotFound)
		return
	}
	out, err := h.redem.Redeem(r.Context(), red, req.WalletAddr, req.Network)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.store.UpdateRedemption(r.Context(), &out); err != nil {
		http.Error(w, "store: "+err.Error(), http.StatusInternalServerError)
		return
	}
	encodeJSON(w, out)
}

func (h *ExchangeHandler) GetRedemption(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("order_id")
	if idStr == "" {
		http.Error(w, "order_id required", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid order_id", http.StatusBadRequest)
		return
	}
	red, err := h.store.GetRedemption(r.Context(), uint(id))
	if err != nil {
		http.Error(w, "redemption not found", http.StatusNotFound)
		return
	}
	encodeJSON(w, red)
}
