package risk

import (
	"context"
	"fmt"

	"alipay-payment/internal/model"
	"alipay-payment/internal/store"
	"github.com/shopspring/decimal"
)

type Engine struct {
	store store.Store
}

type EvaluateRequest struct {
	OrderID           uint
	UserID            uint
	AmountCNY         decimal.Decimal
	IP                string
	DeviceFingerprint string
	GeoCountry        string
	GeoCity           string
	IsVPN             bool
	IsProxy           bool
	IsTor             bool
	Frequency         int
}

type EvaluateResult struct {
	RiskLevel    string
	Score        int
	Action       string
	MatchedRules []MatchedRule
}

type MatchedRule struct {
	RuleID   uint
	RuleName string
	Score    int
	Action   string
}

func NewEngine(s store.Store) *Engine {
	return &Engine{store: s}
}

func (e *Engine) Evaluate(ctx context.Context, req EvaluateRequest) (*EvaluateResult, error) {
	rules, err := e.store.ListRiskRules(ctx)
	if err != nil {
		return nil, err
	}

	result := &EvaluateResult{RiskLevel: "LOW", Score: 0, Action: "ALLOW"}
	totalScore := 0

	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		score, action := e.matchRule(r, req)
		if score > 0 {
			totalScore += score
			result.MatchedRules = append(result.MatchedRules, MatchedRule{
				RuleID: r.ID, RuleName: r.RuleName, Score: score, Action: action,
			})
			if action == "BLOCK" {
				result.Action = "BLOCK"
				result.RiskLevel = "HIGH"
			} else if action == "REVIEW" && result.Action != "BLOCK" {
				result.Action = "REVIEW"
				result.RiskLevel = "MEDIUM"
			}
		}
	}

	result.Score = totalScore
	if totalScore >= 80 {
		result.Action = "BLOCK"
		result.RiskLevel = "CRITICAL"
	} else if totalScore >= 50 {
		result.Action = "REVIEW"
		result.RiskLevel = "HIGH"
	} else if totalScore >= 20 {
		result.RiskLevel = "MEDIUM"
	}

	event := &model.RiskEvent{
		OrderID:           req.OrderID,
		UserID:            req.UserID,
		RiskLevel:         result.RiskLevel,
		Score:             result.Score,
		Action:            result.Action,
		IP:                req.IP,
		DeviceFingerprint: req.DeviceFingerprint,
		GeoCountry:        req.GeoCountry,
		GeoCity:           req.GeoCity,
		IsVPN:             req.IsVPN,
		IsProxy:           req.IsProxy,
		IsTor:             req.IsTor,
	}
	if len(result.MatchedRules) > 0 {
		_ = e.store.CreateRiskEvent(ctx, event)
	}

	return result, nil
}

func (e *Engine) matchRule(r model.RiskRule, req EvaluateRequest) (int, string) {
	switch r.RuleType {
	case "FREQUENCY":
		if req.Frequency > 10 {
			return r.Score, r.Action
		}
	case "AMOUNT":
		if req.AmountCNY.GreaterThan(decimal.NewFromFloat(10000)) {
			return r.Score, r.Action
		}
	case "BLACKLIST":
	case "DEVICE":
		if req.IsVPN || req.IsProxy || req.IsTor {
			return r.Score, r.Action
		}
	case "GEO":
		if req.GeoCountry != "" && req.GeoCountry != "CN" {
			return r.Score, r.Action
		}
	}
	return 0, "ALLOW"
}

var _ = fmt.Sprintf
