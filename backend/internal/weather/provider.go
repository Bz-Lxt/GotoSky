package weather

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gotosky/gotosky/internal/domain"
)

type Provider interface {
	Name() string
	Forecast(ctx context.Context, lat, lon float64, days int) ([]domain.WeatherHour, error)
}

type CallLogFunc func(provider, endpoint string, siteID any, latency time.Duration, code int, cacheHit bool)

type Guard struct {
	mu       sync.Mutex
	byDay    map[string]int
	quota    int
	cache    map[string]cacheEntry
	ttl      time.Duration
	log      CallLogFunc
	inner    Provider
}

type cacheEntry struct {
	at    time.Time
	hours []domain.WeatherHour
}

func NewGuard(inner Provider, quota int, log CallLogFunc) *Guard {
	if quota <= 0 {
		quota = 2000
	}
	return &Guard{
		byDay: make(map[string]int),
		quota: quota,
		cache: make(map[string]cacheEntry),
		ttl:   30 * time.Minute,
		log:   log,
		inner: inner,
	}
}

func (g *Guard) Name() string { return g.inner.Name() }

func (g *Guard) Remaining() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	used := g.byDay[dayKey(time.Now())]
	return g.quota - used
}

func (g *Guard) Forecast(ctx context.Context, lat, lon float64, days int) ([]domain.WeatherHour, error) {
	key := fmt.Sprintf("%.3f,%.3f,%d", lat, lon, days)
	g.mu.Lock()
	if e, ok := g.cache[key]; ok && time.Since(e.at) < g.ttl {
		g.mu.Unlock()
		if g.log != nil {
			g.log(g.inner.Name(), "forecast", nil, 0, 200, true)
		}
		return e.hours, nil
	}
	day := dayKey(time.Now())
	used := g.byDay[day]
	if used >= g.quota {
		g.mu.Unlock()
		return nil, fmt.Errorf("weather quota exceeded")
	}
	g.mu.Unlock()

	start := time.Now()
	hours, err := g.inner.Forecast(ctx, lat, lon, days)
	code := 200
	if err != nil {
		code = 502
	}
	if g.log != nil {
		g.log(g.inner.Name(), "forecast", nil, time.Since(start), code, false)
	}
	if err != nil {
		return nil, err
	}
	g.mu.Lock()
	g.byDay[day] = used + 1
	g.cache[key] = cacheEntry{at: time.Now(), hours: hours}
	g.mu.Unlock()
	return hours, nil
}

func dayKey(t time.Time) string {
	// Quota day is Beijing civil date.
	loc := time.FixedZone("CST", 8*3600)
	y, m, d := t.In(loc).Date()
	return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
}

func KmhToMS(kmh float64) float64 { return kmh / 3.6 }
