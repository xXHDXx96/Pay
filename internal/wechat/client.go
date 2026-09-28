package wechat

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	appID      string
	mchID      string
	serialNo   string
	privateKey *rsa.PrivateKey
	gatewayURL string
	httpClient *http.Client
}

type ClientConfig struct {
	AppID      string
	MchID      string
	SerialNo   string
	PrivateKey string
	GatewayURL string
}

func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.AppID == "" || cfg.MchID == "" || cfg.SerialNo == "" || cfg.PrivateKey == "" {
		return nil, fmt.Errorf("app_id, mch_id, serial_no, private_key required")
	}
	privateKey, err := parsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("private key: %w", err)
	}
	return &Client{
		appID:      cfg.AppID,
		mchID:      cfg.MchID,
		serialNo:   cfg.SerialNo,
		privateKey: privateKey,
		gatewayURL: cfg.GatewayURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (c *Client) signMessage(method, urlPath, timestamp, nonce, body string) (string, error) {
	message := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", method, urlPath, timestamp, nonce, body)
	digest := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(nil, c.privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func (c *Client) authHeader(method, urlPath, body string) (string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())
	sig, err := c.signMessage(method, "/v3/pay/transactions/native", timestamp, nonce, body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`,
		c.mchID, nonce, timestamp, c.serialNo, sig), nil
}

func (c *Client) Precreate(ctx context.Context, outTradeNo, subject, amountCNY string) (string, error) {
	body := fmt.Sprintf(`{"out_trade_no":"%s","description":"%s","amount":{"total":%s,"currency":"CNY"},"notify_url":""}`, outTradeNo, escapeJSON(subject), amountCNY)
	auth, err := c.authHeader("POST", "/v3/pay/transactions/native", body)
	if err != nil {
		return "", err
	}
	req, _ := http.NewRequestWithContext(ctx, "POST", c.gatewayURL+"/v3/pay/transactions/native", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", auth)
	req.Header.Set("Wechatpay-Serial", c.serialNo)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if code, ok := result["code"].(string); ok && code != "SUCCESS" {
		return "", fmt.Errorf("wechat error: %v", result)
	}
	if codeURL, ok := result["code_url"].(string); ok {
		return codeURL, nil
	}
	return "", fmt.Errorf("no code_url in response")
}

func (c *Client) VerifyNotify(headers map[string]string, body []byte) (bool, error) {
	return headers["Wechatpay-Signature"] != "", nil
}

func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
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
