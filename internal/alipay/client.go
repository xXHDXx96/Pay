package alipay

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"alipay-payment/internal/model"
)

type Client struct {
	appID      string
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	gatewayURL string
	httpClient *http.Client
}

type ClientConfig struct {
	AppID           string
	PrivateKeyPEM   string
	AlipayPublicKey string
	GatewayURL      string
}

func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.AppID == "" || cfg.PrivateKeyPEM == "" {
		return nil, fmt.Errorf("app_id and private_key required")
	}
	privateKey, err := parsePrivateKey(cfg.PrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("private key: %w", err)
	}
	var publicKey *rsa.PublicKey
	if cfg.AlipayPublicKey != "" {
		publicKey, err = parsePublicKey(cfg.AlipayPublicKey)
		if err != nil {
			return nil, fmt.Errorf("alipay public key: %w", err)
		}
	}
	return &Client{
		appID:      cfg.AppID,
		privateKey: privateKey,
		publicKey:  publicKey,
		gatewayURL: cfg.GatewayURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (c *Client) signParams(params map[string]string) (string, error) {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString("&")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(params[k])
	}
	digest := sha256.Sum256([]byte(b.String()))
	sig, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func (c *Client) VerifyNotify(form map[string]string) (bool, error) {
	sign := form["sign"]
	if sign == "" {
		return false, fmt.Errorf("missing signature")
	}
	if c.publicKey == nil {
		return false, fmt.Errorf("alipay public key not configured")
	}
	keys := make([]string, 0, len(form))
	for k := range form {
		if k == "sign" || k == "sign_type" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString("&")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(form[k])
	}
	digest := sha256.Sum256([]byte(b.String()))
	sig, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false, fmt.Errorf("bad signature encoding: %w", err)
	}
	return rsa.VerifyPKCS1v15(c.publicKey, crypto.SHA256, digest[:], sig) == nil, nil
}

func (c *Client) Precreate(ctx context.Context, outTradeNo, subject, totalAmount, timeoutExpire string) (string, error) {
	params := map[string]string{
		"app_id":      c.appID,
		"method":      "alipay.trade.precreate",
		"format":      "JSON",
		"charset":     "UTF-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": fmt.Sprintf(`{"out_trade_no":"%s","subject":"%s","total_amount":"%s","timeout_express":"%s"}`, outTradeNo, escapeJSON(subject), totalAmount, timeoutExpire),
	}
	sign, err := c.signParams(params)
	if err != nil {
		return "", err
	}
	params["sign"] = sign

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	req, _ := http.NewRequestWithContext(ctx, "POST", c.gatewayURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("alipay request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var decoded map[string]json.RawMessage
	_ = json.Unmarshal(body, &decoded)
	var biz map[string]string
	const bizKey = "alipay_trade_precreate_response"
	if raw, ok := decoded[bizKey]; ok {
		_ = json.Unmarshal(raw, &biz)
	} else {
		_ = json.Unmarshal(body, &biz)
	}
	if biz["code"] != "10000" {
		return "", fmt.Errorf("alipay error code=%s msg=%s sub=%s", biz["code"], biz["msg"], biz["sub_msg"])
	}
	qr := biz["qr_code"]
	if qr == "" {
		return "", fmt.Errorf("alipay precreate returned empty qr_code")
	}
	return qr, nil
}

func (c *Client) Query(ctx context.Context, outTradeNo string) (map[string]string, error) {
	params := map[string]string{
		"app_id":      c.appID,
		"method":      "alipay.trade.query",
		"format":      "JSON",
		"charset":     "UTF-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": fmt.Sprintf(`{"out_trade_no":"%s"}`, outTradeNo),
	}
	sign, err := c.signParams(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	resp, err := c.httpClient.PostForm(c.gatewayURL, form)
	if err != nil {
		return nil, fmt.Errorf("alipay request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var biz map[string]string
	const bizKey = "alipay_trade_query_response"
	if raw, ok := decodeKey(body, bizKey); ok {
		_ = json.Unmarshal(raw, &biz)
	} else {
		_ = json.Unmarshal(body, &biz)
	}
	return biz, nil
}

func decodeKey(body []byte, key string) ([]byte, bool) {
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, false
	}
	raw, ok := decoded[key]
	return raw, ok
}

func parsePrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		if der, err := base64.StdEncoding.DecodeString(pemData); err == nil {
			block = &pem.Block{Type: "PRIVATE KEY", Bytes: der}
		}
	}
	if block == nil {
		return nil, fmt.Errorf("not a PEM block")
	}
	var key interface{}
	var err error
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	default:
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	}
	if err != nil {
		return nil, err
	}
	pk, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return pk, nil
}

func parsePublicKey(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		if der, err := base64.StdEncoding.DecodeString(pemData); err == nil {
			block = &pem.Block{Type: "PUBLIC KEY", Bytes: der}
		}
	}
	if block == nil {
		return nil, fmt.Errorf("not a PEM block")
	}
	var key interface{}
	var err error
	switch block.Type {
	case "RSA PUBLIC KEY":
		key, err = x509.ParsePKCS1PublicKey(block.Bytes)
	case "PUBLIC KEY":
		key, err = x509.ParsePKIXPublicKey(block.Bytes)
	default:
		key, err = x509.ParsePKIXPublicKey(block.Bytes)
	}
	if err != nil {
		return nil, err
	}
	pk, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return pk, nil
}

func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}

type NotifyPayload struct {
	NotifyTime  string `json:"notify_time"`
	NotifyType  string `json:"notify_type"`
	NotifyID    string `json:"notify_id"`
	TradeNo     string `json:"trade_no"`
	OutTradeNo  string `json:"out_trade_no"`
	TradeStatus string `json:"trade_status"`
	TotalAmount string `json:"total_amount"`
	PayerID     string `json:"buyer_id"`
	Subject     string `json:"subject"`
	PaymentTime string `json:"payment_time"`
	Sign        string `json:"sign"`
}

func ParseNotify(form map[string]string) (*NotifyPayload, error) {
	data, _ := json.Marshal(form)
	var p NotifyPayload
	_ = json.Unmarshal(data, &p)
	p.Sign = form["sign"]
	return &p, nil
}

func (p *NotifyPayload) ToOrder() *model.Order {
	var paidAt *time.Time
	if p.PaymentTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", p.PaymentTime); err == nil {
			paidAt = &t
		}
	}
	return &model.Order{
		TradeNo: p.TradeNo,
		Status:  MapTradeState(p.TradeStatus),
		PaidAt:  paidAt,
		Channel: model.ChannelAlipay,
	}
}

func MapTradeState(tradeStatus string) model.OrderStatus {
	switch tradeStatus {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		return model.OrderPaid
	case "TRADE_CLOSED":
		return model.OrderClosed
	case "TRADE_REFUNDED":
		return model.OrderRefunded
	default:
		return model.OrderPending
	}
}
