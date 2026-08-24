package weather_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gotosky/gotosky/internal/domain"
	"github.com/gotosky/gotosky/internal/weather"
)

func TestGuardForecastCancellationStopsProviderAndPreservesQuota(t *testing.T) {
	provider := &cancellableProvider{
		started:  make(chan struct{}),
		canceled: make(chan struct{}),
		release:  make(chan struct{}),
	}
	guard := weather.NewGuard(provider, 3, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		select {
		case <-provider.release:
		default:
			close(provider.release)
		}
	}()

	done := make(chan error, 1)
	go func() {
		_, err := guard.Forecast(ctx, 31.230, 121.474, 1)
		done <- err
	}()

	select {
	case <-provider.started:
	case <-time.After(time.Second):
		t.Fatal("provider did not start")
	}
	cancel()

	select {
	case <-provider.canceled:
	case <-time.After(200 * time.Millisecond):
		close(provider.release)
	}

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Forecast returned %v after cancellation, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Forecast did not return after cancellation")
	}
	if got := guard.Remaining(); got != 3 {
		t.Errorf("remaining quota = %d after canceled forecast, want 3", got)
	}
}

type cancellableProvider struct {
	started  chan struct{}
	canceled chan struct{}
	release  chan struct{}
}

func (p *cancellableProvider) Name() string { return "cancellable" }

func (p *cancellableProvider) Forecast(ctx context.Context, _, _ float64, _ int) ([]domain.WeatherHour, error) {
	close(p.started)
	select {
	case <-ctx.Done():
		close(p.canceled)
		return nil, ctx.Err()
	case <-p.release:
		if err := ctx.Err(); err != nil {
			close(p.canceled)
			return nil, err
		}
		return []domain.WeatherHour{{}}, nil
	}
}
