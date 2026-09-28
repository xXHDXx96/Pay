package reconciliation

import (
	"context"
	"fmt"
	"time"

	"alipay-payment/internal/model"
	"alipay-payment/internal/store"
	"github.com/shopspring/decimal"
)

type RemoteStatement struct {
	OrderNo   string
	AmountCNY float64
	Status    string
	PaidAt    time.Time
}

type Provider interface {
	FetchStatement(ctx context.Context, date time.Time) ([]RemoteStatement, error)
}

type Service struct {
	store     store.Store
	providers map[string]Provider
	stopCh    chan struct{}
}

func NewService(s store.Store) *Service {
	return &Service{store: s, providers: make(map[string]Provider), stopCh: make(chan struct{})}
}

func (s *Service) RegisterProvider(name string, p Provider) { s.providers[name] = p }

func (s *Service) Scheduler() *Scheduler {
	return &Scheduler{svc: s, stopCh: make(chan struct{})}
}

type Scheduler struct {
	svc    *Service
	stopCh chan struct{}
}

func (sc *Scheduler) Start() {
	go sc.run()
}
func (sc *Scheduler) Stop() { close(sc.stopCh) }
func (sc *Scheduler) run() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-sc.stopCh:
			return
		case <-ticker.C:
			_ = sc.svc.runDaily()
		}
	}
}

func (s *Service) runDaily() error {
	ctx := context.Background()
	date := time.Now().AddDate(0, 0, -1)
	for name, p := range s.providers {
		remotes, err := p.FetchStatement(ctx, date)
		if err != nil {
			return err
		}
		_, err = s.Reconcile(ctx, name, date, remotes)
		return err
	}
	return nil
}

func (s *Service) Reconcile(ctx context.Context, channel string, date time.Time, remotes []RemoteStatement) (*model.Reconciliation, error) {
	recon := &model.Reconciliation{
		Date:        date,
		Channel:     channel,
		Provider:    channel,
		TotalRemote: len(remotes),
		Status:      "PENDING",
	}
	if err := s.store.CreateReconciliation(ctx, recon); err != nil {
		return nil, err
	}

	localOrders, err := s.store.ListOrders(ctx, "")
	if err != nil {
		return nil, err
	}

	remoteMap := make(map[string]RemoteStatement)
	for _, r := range remotes {
		remoteMap[r.OrderNo] = r
	}

	matched := 0
	var discrepancies []model.ReconciliationDetail

	for _, lo := range localOrders {
		remote, ok := remoteMap[lo.OutOrderNo]
		if !ok {
			detail := &model.ReconciliationDetail{
				ReconID:      recon.ID,
				LocalOrderNo: lo.OutOrderNo,
				LocalAmount:  lo.AmountCNY,
				LocalStatus:  string(lo.Status),
				Status:       "MISSING_REMOTE",
				Discrepancy:  fmt.Sprintf(`{"reason":"remote order not found"}`),
			}
			_ = s.store.CreateReconciliationDetail(ctx, detail)
			discrepancies = append(discrepancies, *detail)
			continue
		}
		if remote.Status == "PAID" && lo.Status == "PAID" {
			matched++
			detail := &model.ReconciliationDetail{
				ReconID:       recon.ID,
				LocalOrderNo:  lo.OutOrderNo,
				RemoteOrderNo: remote.OrderNo,
				LocalAmount:   lo.AmountCNY,
				RemoteAmount:  decimal.NewFromFloat(remote.AmountCNY),
				LocalStatus:   string(lo.Status),
				RemoteStatus:  remote.Status,
				Status:        "MATCHED",
				Discrepancy:   "{}",
			}
			_ = s.store.CreateReconciliationDetail(ctx, detail)
		} else {
			detail := &model.ReconciliationDetail{
				ReconID:       recon.ID,
				LocalOrderNo:  lo.OutOrderNo,
				RemoteOrderNo: remote.OrderNo,
				LocalAmount:   lo.AmountCNY,
				RemoteAmount:  decimal.NewFromFloat(remote.AmountCNY),
				LocalStatus:   string(lo.Status),
				RemoteStatus:  remote.Status,
				Status:        "MISMATCH",
				Discrepancy:   fmt.Sprintf(`{"local_status":"%s","remote_status":"%s"}`, lo.Status, remote.Status),
			}
			_ = s.store.CreateReconciliationDetail(ctx, detail)
			discrepancies = append(discrepancies, *detail)
		}
	}

	for _, r := range remotes {
		found := false
		for _, lo := range localOrders {
			if lo.OutOrderNo == r.OrderNo {
				found = true
				break
			}
		}
		if !found {
			detail := &model.ReconciliationDetail{
				ReconID:       recon.ID,
				RemoteOrderNo: r.OrderNo,
				RemoteAmount:  decimal.NewFromFloat(r.AmountCNY),
				RemoteStatus:  r.Status,
				Status:        "MISSING_LOCAL",
				Discrepancy:   fmt.Sprintf(`{"reason":"local order not found"}`),
			}
			_ = s.store.CreateReconciliationDetail(ctx, detail)
			discrepancies = append(discrepancies, *detail)
		}
	}

	recon.MatchedCount = matched
	recon.MismatchCount = len(discrepancies)
	recon.MissingLocal = 0
	recon.MissingRemote = 0
	if len(discrepancies) == 0 {
		recon.Status = "MATCHED"
	} else {
		recon.Status = "MISMATCH"
	}
	now := time.Now()
	recon.CompletedAt = &now
	_ = s.store.UpdateReconciliation(ctx, recon)

	return recon, nil
}

func (s *Service) ListDiscrepancies(ctx context.Context, reconID uint) ([]model.ReconciliationDetail, error) {
	return s.store.ListReconciliationDetails(ctx, reconID)
}

func (s *Service) ResolveDiscrepancy(ctx context.Context, detailID uint, resolvedBy uint, note string) error {
	details, err := s.store.ListReconciliationDetails(ctx, 0)
	if err != nil {
		return err
	}
	for _, d := range details {
		if d.ID == detailID {
			d.Resolved = true
			d.ResolvedBy = resolvedBy
			d.ResolutionNote = note
			now := time.Now()
			d.ResolvedAt = &now
			return s.store.UpdateReconciliationDetail(ctx, &d)
		}
	}
	return fmt.Errorf("detail not found: %d", detailID)
}
