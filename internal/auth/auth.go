package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"alipay-payment/internal/store"
)

var (
	ErrInvalidAPIKey    = errors.New("invalid api key")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrMerchantInactive = errors.New("merchant inactive")
)

type Authenticator struct {
	store store.Store
}

func NewAuthenticator(s store.Store) *Authenticator {
	return &Authenticator{store: s}
}

func (a *Authenticator) Authenticate(r *http.Request) (uint, error) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		apiKey = r.URL.Query().Get("api_key")
	}
	if apiKey == "" {
		return 0, ErrInvalidAPIKey
	}

	m, err := a.store.GetMerchant(r.Context(), apiKey)
	if err != nil {
		return 0, ErrInvalidAPIKey
	}
	if !m.Active {
		return 0, ErrMerchantInactive
	}

	if m.WebhookSecret != "" {
		sig := r.Header.Get("X-Signature")
		if sig == "" {
			return 0, ErrInvalidSignature
		}
		if !a.verifyHMAC(m.WebhookSecret, sig, r) {
			return 0, ErrInvalidSignature
		}
	}

	return m.ID, nil
}

func (a *Authenticator) verifyHMAC(secret, got string, r *http.Request) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(r.Method))
	mac.Write([]byte(r.URL.Path))
	if r.URL.RawQuery != "" {
		mac.Write([]byte("?"))
		mac.Write([]byte(r.URL.RawQuery))
	}
	want := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(got), []byte(want))
}

func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := a.Authenticate(r)
		if err != nil {
			if errors.Is(err, ErrInvalidAPIKey) {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			} else {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GenerateAPIKey() string {
	b := make([]byte, 24)
	rand.Read(b)
	return "pk_" + hex.EncodeToString(b)
}

func GenerateSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

var _ = strings.TrimSpace
