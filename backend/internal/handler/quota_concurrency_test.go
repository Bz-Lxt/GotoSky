package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gotosky/gotosky/internal/config"
	"github.com/gotosky/gotosky/internal/domain"
	"github.com/gotosky/gotosky/internal/handler"
	"github.com/gotosky/gotosky/internal/weather"
	"github.com/gotosky/gotosky/internal/ws"
)

type quotaBarrierProvider struct {
	mu      sync.Mutex
	want    int
	calls   int
	ready   chan struct{}
	release chan struct{}
}

func newQuotaBarrierProvider(want int) *quotaBarrierProvider {
	return &quotaBarrierProvider{
		want:    want,
		ready:   make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (p *quotaBarrierProvider) Name() string { return "barrier" }

func (p *quotaBarrierProvider) Forecast(ctx context.Context, _, _ float64, _ int) ([]domain.WeatherHour, error) {
	p.mu.Lock()
	p.calls++
	if p.calls == p.want {
		close(p.ready)
	}
	p.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-p.release:
		return []domain.WeatherHour{{TimeUTC: time.Now().UTC()}}, nil
	}
}

func TestQuotaReportsAllConcurrentProviderCalls(t *testing.T) {
	const (
		callCount  = 4
		dailyQuota = 10
	)
	provider := newQuotaBarrierProvider(callCount)
	guard := weather.NewGuard(provider, dailyQuota, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var releaseOnce sync.Once
	release := func() {
		releaseOnce.Do(func() { close(provider.release) })
	}
	defer release()

	errs := make(chan error, callCount)
	for i := 0; i < callCount; i++ {
		go func(i int) {
			_, err := guard.Forecast(ctx, float64(i), float64(i), 1)
			errs <- err
		}(i)
	}

	select {
	case <-provider.ready:
	case <-ctx.Done():
		t.Fatalf("provider calls did not overlap: %v", ctx.Err())
	}
	release()
	for i := 0; i < callCount; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("forecast %d failed: %v", i, err)
		}
	}

	api := &handler.API{
		Cfg:   config.Config{WeatherQuota: dailyQuota},
		Guard: guard,
		Hub:   ws.NewHub(nil, ""),
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/quota", nil)
	api.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/quota status = %d, want %d", rec.Code, http.StatusOK)
	}
	var response struct {
		Data struct {
			Remaining int `json:"remaining"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode quota response: %v", err)
	}
	if want := dailyQuota - callCount; response.Data.Remaining != want {
		t.Fatalf("remaining quota = %d, want %d after %d provider calls", response.Data.Remaining, want, callCount)
	}
}
