package weather

import (
	"context"
	"math"
	"time"

	"github.com/gotosky/gotosky/internal/domain"
)

// Mock is deterministic: same seed hour → same values. No DNS.
type Mock struct {
	Script string // CLEAR | MIXED | CLOUDY | STORM
}

func NewMock(script string) *Mock {
	if script == "" {
		script = "MIXED"
	}
	return &Mock{Script: script}
}

func (m *Mock) Name() string { return "mock" }

func (m *Mock) Forecast(_ context.Context, lat, lon float64, days int) ([]domain.WeatherHour, error) {
	if days < 1 {
		days = 7
	}
	start := time.Now().UTC().Truncate(time.Hour)
	n := days * 24
	out := make([]domain.WeatherHour, 0, n)
	for i := 0; i < n; i++ {
		t := start.Add(time.Duration(i) * time.Hour)
		h := t.Hour()
		phase := float64(i) / 24
		switch m.Script {
		case "CLEAR":
			out = append(out, hour(t, 4, 6, 8, 38, 12, 4, 25000, 0, 2.2, 3.0, 21.5, 12.5, 2.4))
		case "CLOUDY":
			out = append(out, hour(t, 70, 60, 40, 88, 18, 16, 4000, 20, 4, 6, 35, 20, 8))
		case "STORM":
			out = append(out, hour(t, 90, 80, 70, 96, 20, 19.5, 2000, 80, 8, 14, 45, 30, 12))
		default: // MIXED: good nights around 21-02 local-ish UTC+8 → UTC 13-18
			cloud := 12 + 40*math.Abs(math.Sin(phase+lat/90))
			if h >= 13 && h <= 18 {
				cloud = 8 + float64(i%3)*2
			}
			if i%36 == 10 {
				cloud = 72
			}
			rh := 45 + 20*math.Sin(phase+lon/180)
			out = append(out, hour(t, cloud, cloud*0.7, cloud*0.4, rh, 10, 4, 18000, float64(i%20), 2.2+float64(i%5)*0.2, 3.5, 21.5, 12.5, 2.4))
		}
	}
	return out, nil
}

func hour(t time.Time, cl, cm, ch, rh, temp, dew, vis, pp, w10, gust, w250, w500, w850 float64) domain.WeatherHour {
	return domain.WeatherHour{
		TimeUTC: t, CloudLow: cl, CloudMid: cm, CloudHigh: ch,
		RH: rh, TempC: temp, DewC: dew, VisibilityM: vis, PrecipProb: pp,
		Wind10MS: w10, Gust10MS: gust, Wind250MS: w250, Wind500MS: w500, Wind850MS: w850,
	}
}
