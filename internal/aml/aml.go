package aml

import (
	"context"
	"fmt"
	"time"

	"alipay-payment/internal/model"
	"alipay-payment/internal/store"
	"github.com/shopspring/decimal"
)

type Engine struct {
	store store.Store
}

type ScreeningRequest struct {
	UserID    uint
	Name      string
	IDCard    string
	Country   string
	AmountCNY decimal.Decimal
}

type ScreeningResult struct {
	Status    string
	Score     decimal.Decimal
	RiskLevel string
	Matches   []string
}

func NewEngine(s store.Store) *Engine {
	return &Engine{store: s}
}

func (e *Engine) Screen(ctx context.Context, req ScreeningRequest) (*ScreeningResult, error) {
	result := &ScreeningResult{
		Status:    "PASS",
		Score:     decimal.NewFromFloat(0.00),
		RiskLevel: "LOW",
		Matches:   []string{},
	}

	if req.Country != "" && req.Country != "CN" {
		result.Matches = append(result.Matches, fmt.Sprintf("GEO_RISK:%s", req.Country))
		result.Score = result.Score.Add(decimal.NewFromFloat(0.3))
		result.RiskLevel = "MEDIUM"
	}

	if req.AmountCNY.GreaterThan(decimal.NewFromFloat(50000)) {
		result.Matches = append(result.Matches, "LARGE_TRANSACTION")
		result.Score = result.Score.Add(decimal.NewFromFloat(0.2))
		if result.RiskLevel == "LOW" {
			result.RiskLevel = "MEDIUM"
		}
	}

	if result.Score.GreaterThanOrEqual(decimal.NewFromFloat(0.6)) {
		result.Status = "REVIEW"
		result.RiskLevel = "HIGH"
	} else if result.Score.GreaterThan(decimal.NewFromFloat(0.3)) {
		result.Status = "REVIEW"
	} else {
		result.Status = "PASS"
	}

	_ = e.store.CreateAML(ctx, &model.AMLScreening{
		UserID: req.UserID, Provider: "INTERNAL", Status: result.Status,
		Score: result.Score, Matches: fmt.Sprintf("%v", result.Matches), RiskLevel: result.RiskLevel,
	})

	if result.Status == "REVIEW" {
		_ = e.store.CreateAuditLog(ctx, &model.AuditLog{
			Action: "AML_ALERT", ResourceType: "AML", ResourceID: fmt.Sprintf("user-%d", req.UserID),
			Severity: "WARN", ChangeSummary: fmt.Sprintf("AML screening flagged: %v", result.Matches),
		})
	}

	return result, nil
}

func (e *Engine) VerifyKYC(ctx context.Context, userID uint, name, idCard string) (*model.KYCVerification, error) {
	kyc := &model.KYCVerification{
		UserID: userID, Provider: "INTERNAL", Status: "PENDING",
		VerifiedName: name, VerifiedIDCard: idCard,
	}
	if err := e.store.CreateKYC(ctx, kyc); err != nil {
		return nil, err
	}

	kyc.Status = "APPROVED"
	now := time.Now()
	kyc.ExpiresAt = &now
	*kyc.ExpiresAt = now.AddDate(2, 0, 0)
	_ = e.store.UpdateKYC(ctx, kyc)

	return kyc, nil
}
