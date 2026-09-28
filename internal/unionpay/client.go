package unionpay

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type ClientConfig struct {
	MchID      string
	PublicKey  string
	PrivateKey string
	GatewayURL string
	Name       string
}

type Client struct {
	mchID      string
	name       string
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	gatewayURL string
	httpClient *http.Client
}

type QRPayParams struct {
	OutTradeNo string
	Subject    string
	AmountCNY  string
	IP         string
}

type NotifyParams map[string]string

func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.MchID == "" {
		return nil, fmt.Errorf("mch_id required")
	}
	pubKey, err := parsePublicKey(cfg.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("public key: %w", err)
	}
	privKey, err := parsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("private key: %w", err)
	}
	return &Client{
		mchID:      cfg.MchID,
		name:       cfg.Name,
		publicKey:  pubKey,
		privateKey: privKey,
		gatewayURL: strings.TrimSuffix(cfg.GatewayURL, "/"),
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (c *Client) Precreate(ctx context.Context, outTradeNo, subject, amountCNY string) (string, error) {
	params := map[string]string{
		"version":      "1.0.0",
		"encoding":     "UTF-8",
		"signType":     "RSA2",
		"certType":     "",
		"mercdeId":     c.mchID,
		"orderId":      outTradeNo,
		"txnTime":      time.Now().Format("20060102150405"),
		"txnAmt":       mul100(amountCNY),
		"txnType":      "01",
		"txnSubType":   "01",
		"bizType":      "000000",
		"accessType":   "00",
		"channelType":  "07",
		"terrName":     "01",
		"merName":      c.name,
		"orderDesc":    subject,
		"backUrl":      "",
		"frontUrl":     "",
		"autoLoadTime": "60",
	}

	signedParams := signParams(params, c.privateKey)

	form := url.Values{}
	for k, v := range signedParams {
		form.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.gatewayURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("unionpay request failed: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if result["respCode"] != "00" {
		return "", fmt.Errorf("unionpay error: %v", result["respMsg"])
	}

	if codeURL, ok := result["qrCode"].(string); ok {
		return codeURL, nil
	}
	return "", fmt.Errorf("no qr_code in response")
}

func (c *Client) VerifyNotify(params map[string]string) (bool, error) {
	sign := params["signature"]
	if sign == "" {
		return false, fmt.Errorf("signature missing")
	}

	delete(params, "signature")
	delete(params, "signType")

	sortedKeys := make([]string, 0, len(params))
	for k := range params {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	buf := strings.Builder{}
	for _, k := range sortedKeys {
		if params[k] != "" {
			buf.WriteString(k)
			buf.WriteString("=")
			buf.WriteString(params[k])
			buf.WriteString("&")
		}
	}
	content := strings.TrimSuffix(buf.String(), "&")

	sigBytes, err := hex.DecodeString(sign)
	if err != nil {
		return false, fmt.Errorf("decode signature: %w", err)
	}

	hashed := sha256.Sum256([]byte(content))
	if err := rsa.VerifyPKCS1v15(c.publicKey, crypto.SHA256, hashed[:], sigBytes); err != nil {
		return false, fmt.Errorf("signature verification failed: %w", err)
	}

	return true, nil
}

func signParams(params map[string]string, privKey *rsa.PrivateKey) map[string]string {
	sortedKeys := make([]string, 0, len(params))
	for k := range params {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	buf := strings.Builder{}
	for _, k := range sortedKeys {
		if params[k] != "" {
			buf.WriteString(k)
			buf.WriteString("=")
			buf.WriteString(params[k])
			buf.WriteString("&")
		}
	}
	content := strings.TrimSuffix(buf.String(), "&")

	hashed := sha256.Sum256([]byte(content))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
	if err != nil {
		return params
	}

	params["signature"] = hex.EncodeToString(sig)
	return params
}

func mul100(amountCNY string) string {
	f, err := parseDecimal(amountCNY)
	if err != nil {
		return "0"
	}
	result := f.Mul(f, big.NewFloat(100))
	return result.String()
}

func parseDecimal(s string) (*big.Float, error) {
	f, _, err := big.ParseFloat(s, 10, 0, big.ToNearestEven)
	return f, err
}

func parsePublicKey(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("not a PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaPub, nil
}

func parsePrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("not a PEM block")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}
	key2, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	rsaKey, ok := key2.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return rsaKey, nil
}
