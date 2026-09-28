package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"alipay-payment/internal/alipay"
	"alipay-payment/internal/config"
	"alipay-payment/internal/exchange"
	"alipay-payment/internal/handler"
	"alipay-payment/internal/manualpay"
	"alipay-payment/internal/model"
	"alipay-payment/internal/store"
	"alipay-payment/internal/unionpay"
	"alipay-payment/internal/wechat"
)

func main() {
	cfg := config.Load()

	var alipayQR func(outOrderNo, subject string, amountCNY float64) (string, error)
	var alipayNotify func(map[string]string) (bool, error)
	if cfg.AlipayAppID != "" && cfg.AlipayPrivateKey != "" {
		cli, err := alipay.NewClient(alipay.ClientConfig{
			AppID:           cfg.AlipayAppID,
			PrivateKeyPEM:   cfg.AlipayPrivateKey,
			AlipayPublicKey: cfg.AlipayPublicKey,
			GatewayURL:      cfg.AlipayGatewayURL,
		})
		if err != nil {
			log.Fatalf("alipay client: %v", err)
		}
		alipayNotify = cli.VerifyNotify
		alipayQR = func(outOrderNo, subject string, amountCNY float64) (string, error) {
			return cli.Precreate(context.Background(), outOrderNo, subject, fmt.Sprintf("%.2f", amountCNY), "5m")
		}
	}

	var wechatQR func(outOrderNo, subject string, amountCNY float64) (string, error)
	if cfg.WechatAppID != "" && cfg.WechatMchID != "" && cfg.WechatSerialNo != "" && cfg.WechatPrivateKey != "" {
		cli, err := wechat.NewClient(wechat.ClientConfig{
			AppID:      cfg.WechatAppID,
			MchID:      cfg.WechatMchID,
			SerialNo:   cfg.WechatSerialNo,
			PrivateKey: cfg.WechatPrivateKey,
			GatewayURL: cfg.WechatGatewayURL,
		})
		if err != nil {
			log.Printf("wechat client: %v", err)
		} else {
			wechatQR = func(outOrderNo, subject string, amountCNY float64) (string, error) {
				return cli.Precreate(context.Background(), outOrderNo, subject, fmt.Sprintf("%.2f", amountCNY))
			}
		}
	}

	var unionPayQR func(outOrderNo, subject string, amountCNY float64) (string, error)
	if cfg.UnionPayMchID != "" && cfg.UnionPayPublicKey != "" && cfg.UnionPayPrivateKey != "" {
		unionPayClient, err := unionpay.NewClient(unionpay.ClientConfig{
			MchID:      cfg.UnionPayMchID,
			PublicKey:  cfg.UnionPayPublicKey,
			PrivateKey: cfg.UnionPayPrivateKey,
			GatewayURL: cfg.UnionPayGatewayURL,
			Name:       cfg.UnionPayName,
		})
		if err != nil {
			log.Printf("unionpay client: %v", err)
		} else {
			unionPayQR = func(outOrderNo, subject string, amountCNY float64) (string, error) {
				return unionPayClient.Precreate(context.Background(), outOrderNo, subject, fmt.Sprintf("%.2f", amountCNY))
			}
		}
	}

	// ============ 数据库初始化（原内存存储替换为 PostgreSQL） ============
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("FATAL: 环境变量 DATABASE_URL 未设置。请在 Render 控制台添加 Postgres 数据库，并填入其 Internal Database URL。")
	}
	var s store.Store
	s, err := store.NewPostgresStore(dbURL)
	if err != nil {
		log.Fatalf("FATAL: 连接 PostgreSQL 数据库失败: %v", err)
	}
	log.Println("成功连接到 PostgreSQL 数据库")
	// ===================================================================

	oracle := exchange.NewOracle()
	redem := exchange.NewRedemptions(oracle)

	payHandler := handler.NewPayHandler(s).
		WithAlipayQR(alipayQR).
		WithWechatQR(wechatQR).
		WithUnionPayQR(unionPayQR)

	qrHandler := handler.NewQRCodeHandler(s)
	if alipayQR != nil {
		qrHandler = qrHandler.WithChannel(model.ChannelAlipay, alipayQR)
	}
	if wechatQR != nil {
		qrHandler = qrHandler.WithChannel(model.ChannelWechat, wechatQR)
	}
	if unionPayQR != nil {
		qrHandler = qrHandler.WithChannel(model.ChannelUnionpay, unionPayQR)
	}

	notifyHandler := handler.NewNotifyHandler(s, alipayNotify)
	exchangeHandler := handler.NewExchangeHandler(s, redem)
	webhookHandler := handler.NewWebhookHandler(s, cfg.WebhookSecret)
	manualPayHandler := manualpay.NewHandler(s, redem)

	mux := http.NewServeMux()

	// 原有路由
	mux.Handle("/api/v1/pay", payHandler)
	mux.Handle("/api/v1/orders", handler.NewOrdersHandler(s))
	mux.Handle("/api/v1/orders/", handler.NewOrderHandler(s))
	mux.Handle("/api/v1/notify/alipay", notifyHandler)
	mux.Handle("/api/v1/notify/wechat", notifyHandler)
	mux.Handle("/api/v1/qrcode", qrHandler)
	mux.Handle("/api/v1/qrcode/", qrHandler)
	mux.Handle("/api/v1/webhook", webhookHandler)
	mux.Handle("/api/v1/exchange/quote", http.HandlerFunc(exchangeHandler.Quote))
	mux.Handle("/api/v1/exchange/redeem", http.HandlerFunc(exchangeHandler.Redeem))
	mux.Handle("/api/v1/exchange", http.HandlerFunc(exchangeHandler.GetRedemption))
	mux.Handle("/health", handler.NewHealthHandler())

	// 免签支付路由
	mux.HandleFunc("/api/v1/manual/create", manualPayHandler.CreateOrder)
	mux.HandleFunc("/api/v1/manual/notify", manualPayHandler.NotifyFromPhone)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		log.Printf("payment gateway listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	_ = srv.Close()
	log.Println("server stopped")
}
