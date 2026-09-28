package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"net/http"

	"alipay-payment/internal/alipay"
	"alipay-payment/internal/store"
)

type NotifyHandler struct {
	store    store.Store
	verifier func(map[string]string) (bool, error)
	secret   string
}

func NewNotifyHandler(s store.Store, verifier func(map[string]string) (bool, error)) *NotifyHandler {
	return &NotifyHandler{store: s, verifier: verifier}
}

func (h *NotifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	form := make(map[string]string)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form: "+err.Error(), http.StatusBadRequest)
		return
	}
	for k, v := range r.Form {
		if len(v) > 0 {
			form[k] = v[0]
		}
	}

	if h.secret != "" {
		got := r.Header.Get("X-Signature")
		if got == "" || !h.verifySecret(got, form) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	ok, err := h.verifier(form)
	if err != nil || !ok {
		http.Error(w, "signature verification failed", http.StatusUnauthorized)
		return
	}

	notify, err := alipay.ParseNotify(form)
	if err != nil {
		http.Error(w, "notify parse: "+err.Error(), http.StatusBadRequest)
		return
	}

	o, err := h.store.GetOrderByOutNo(r.Context(), notify.OutTradeNo)
	if err != nil {
		w.Write([]byte("success"))
		return
	}

	updated := notify.ToOrder()
	updated.ID = o.ID
	updated.MerchantID = o.MerchantID
	updated.OutOrderNo = o.OutOrderNo
	updated.Subject = o.Subject
	updated.AmountCNY = o.AmountCNY
	updated.Channel = o.Channel
	if err := h.store.UpdateOrder(r.Context(), updated); err != nil {
		http.Error(w, "update order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("success"))
}

func (h *NotifyHandler) verifySecret(got string, form map[string]string) bool {
	mac := hmac.New(sha256.New, []byte(h.secret))
	for k := range form {
		if k == "sign" || k == "sign_type" {
			continue
		}
		mac.Write([]byte(k))
		mac.Write([]byte(form[k]))
	}
	want := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(got), []byte(want))
}
