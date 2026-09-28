package manualpay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"alipay-payment/internal/exchange"
	"alipay-payment/internal/model"
	"alipay-payment/internal/store"

	"github.com/shopspring/decimal"
)

type Handler struct {
	store  store.Store
	redeem *exchange.Redemptions
}

func NewHandler(s store.Store, r *exchange.Redemptions) *Handler {
	return &Handler{
		store:  s,
		redeem: r,
	}
}

// CreateOrder 分配唯一金额，创建免签订单
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Amount     float64 `json:"amount"`
		MerchantID string  `json:"merchant_id"`
		WalletAddr string  `json:"wallet_address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	baseAmount := decimal.NewFromFloat(req.Amount).Round(2)

	var candidates []decimal.Decimal
	candidates = append(candidates, baseAmount)
	for i := 1; i <= 10; i++ {
		delta := decimal.NewFromFloat(0.01 * float64(i))
		if baseAmount.Sub(delta).IsPositive() {
			candidates = append(candidates, baseAmount.Sub(delta))
		}
		candidates = append(candidates, baseAmount.Add(delta))
	}

	var finalOrder *model.Order
	for _, amount := range candidates {
		order := &model.Order{
			MerchantID: req.MerchantID,
			OutOrderNo: fmt.Sprintf("MAN%d%04d", time.Now().UnixNano()/1e6, rand.Intn(10000)),
			Subject:    "免签支付",
			AmountCNY:  amount,
			Channel:    model.ChannelAlipay,
			Status:     model.OrderPending,
		}

		err := h.store.CreateOrder(context.Background(), order)
		if err == nil {
			finalOrder = order
			break
		}
		if err.Error() == "AMOUNT_OCCUPIED" {
			continue
		}
		http.Error(w, "系统内部错误", 500)
		return
	}

	if finalOrder == nil {
		http.Error(w, "当前支付排队人数过多，请稍后重试", 429)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":          200,
		"order_no":      finalOrder.OutOrderNo,
		"actual_amount": finalOrder.AmountCNY.StringFixed(2),
	})
}

// NotifyFromPhone 接收挂机手机发来的到账通知
func (h *Handler) NotifyFromPhone(w http.ResponseWriter, r *http.Request) {
	secret := r.Header.Get("X-Notify-Secret")
	if secret != os.Getenv("NOTIFY_SECRET") {
		http.Error(w, "Unauthorized", 401)
		return
	}

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	amount := decimal.NewFromFloat(req.Amount).Round(2)

	if err := h.store.MarkOrderPaid(context.Background(), amount); err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("no match or already paid"))
		return
	}

	go func() {
		log.Printf("[自动兑换] 订单金额 %s 已到账，等待 App 确认兑换...", amount.StringFixed(2))
	}()

	w.Write([]byte("success"))
}