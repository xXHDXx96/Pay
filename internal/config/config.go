package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port             string
	AlipayAppID      string
	AlipayPrivateKey string
	AlipayPublicKey  string
	AlipayRootCert   string
	AlipayGatewayURL string
	NotifyBaseURL    string

	WechatAppID      string
	WechatMchID      string
	WechatSerialNo   string
	WechatPrivateKey string
	WechatCertPath   string
	WechatGatewayURL string

	UnionPayMchID      string
	UnionPayPublicKey  string
	UnionPayPrivateKey string
	UnionPayGatewayURL string
	UnionPayName       string

	EthereumRPC          string
	CryptoMerchantWallet string
	USDTContract         string

	DatabaseURL   string
	WebhookSecret string

	RedisURL     string
	RateLimitRPS int
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func Load() Config {
	return Config{
		Port:             env("PORT", "8080"),
		AlipayAppID:      env("ALIPAY_APP_ID", ""),
		AlipayPrivateKey: env("ALIPAY_PRIVATE_KEY", ""),
		AlipayPublicKey:  env("ALIPAY_PUBLIC_KEY", ""),
		AlipayRootCert:   env("ALIPAY_ROOT_CERT", ""),
		AlipayGatewayURL: env("ALIPAY_GATEWAY_URL", "https://openapi.alipay.com/gateway.do"),
		NotifyBaseURL:    env("NOTIFY_BASE_URL", "http://localhost:8080"),
		WechatAppID:      env("WECHAT_APP_ID", ""),
		WechatMchID:      env("WECHAT_MCH_ID", ""),
		WechatSerialNo:   env("WECHAT_SERIAL_NO", ""),
		WechatPrivateKey: env("WECHAT_PRIVATE_KEY", ""),
		WechatCertPath:   env("WECHAT_CERT_PATH", ""),
		WechatGatewayURL: env("WECHAT_GATEWAY_URL", "https://api.mch.weixin.qq.com"),

		UnionPayMchID:        env("UNIONPAY_MCH_ID", ""),
		UnionPayPublicKey:    env("UNIONPAY_PUBLIC_KEY", ""),
		UnionPayPrivateKey:   env("UNIONPAY_PRIVATE_KEY", ""),
		UnionPayGatewayURL:   env("UNIONPAY_GATEWAY_URL", "https://gateway.9pay.cn"),
		UnionPayName:         env("UNIONPAY_NAME", ""),
		EthereumRPC:          env("ETH_RPC", ""),
		CryptoMerchantWallet: env("CRYPTO_MERCHANT_WALLET", ""),
		USDTContract:         env("USDT_CONTRACT", "0xdAC17F958D2ee523a2206206994597C13D831ec7"),
		DatabaseURL:          env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/payments?sslmode=disable"),
		WebhookSecret:        env("WEBHOOK_SECRET", "dev-secret"),
		RedisURL:             env("REDIS_URL", ""),
		RateLimitRPS:         envInt("RATE_LIMIT_RPS", 100),
	}
}
