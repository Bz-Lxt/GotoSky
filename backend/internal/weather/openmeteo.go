package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gotosky/gotosky/internal/domain"
)

const omURL = "https://api.open-meteo.com/v1/forecast"

type OpenMeteo struct {
	Client *http.Client
}

func NewOpenMeteo() *OpenMeteo {
	return &OpenMeteo{Client: &http.Client{Timeout: 12 * time.Second}}
}

func (o *OpenMeteo) Name() string { return "openmeteo" }

type omResp struct {
	Error       bool   `json:"error"`
	Reason      string `json:"reason"`
	HourlyUnits map[string]string `json:"hourly_units"`
	Hourly      struct {
		Time                     []string   `json:"time"`
		CloudLow                 []float64  `json:"cloud_cover_low"`
		CloudMid                 []float64  `json:"cloud_cover_mid"`
		CloudHigh                []float64  `json:"cloud_cover_high"`
		RH                       []float64  `json:"relative_humidity_2m"`
		Temp                     []float64  `json:"temperature_2m"`
		Dew                      []float64  `json:"dew_point_2m"`
		Vis                      []float64  `json:"visibility"`
		Precip                   []float64  `json:"precipitation_probability"`
		Wind10                   []float64  `json:"wind_speed_10m"`
		Gust                     []float64  `json:"wind_gusts_10m"`
		W250                     []float64  `json:"wind_speed_250hPa"`
		W500                     []float64  `json:"wind_speed_500hPa"`
		W850                     []float64  `json:"wind_speed_850hPa"`
	} `json:"hourly"`
}

func (o *OpenMeteo) Forecast(ctx context.Context, lat, lon float64, days int) ([]domain.WeatherHour, error) {
	if days < 1 {
		days = 7
	}
	if days > 8 {
		days = 8
	}
	q := fmt.Sprintf("%s?latitude=%f&longitude=%f&forecast_days=%d&timezone=UTC&hourly=cloud_cover_low,cloud_cover_mid,cloud_cover_high,relative_humidity_2m,dew_point_2m,temperature_2m,visibility,precipitation_probability,wind_speed_10m,wind_gusts_10m,wind_speed_250hPa,wind_speed_500hPa,wind_speed_850hPa",
		omURL, lat, lon, days)
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, q, nil)
		if err != nil {
			return nil, err
		}
		resp, err := o.Client.Do(req)
		if err != nil {
			last = err
			if ctx.Err() != nil {
				return nil, err
			}
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == 400 || resp.StatusCode == 401 || resp.StatusCode == 422 {
			return nil, fmt.Errorf("open-meteo validation: %s", body)
		}
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			last = fmt.Errorf("open-meteo http %d", resp.StatusCode)
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			continue
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("open-meteo http %d", resp.StatusCode)
		}
		var parsed omResp
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, fmt.Errorf("open-meteo json: %w", err)
		}
		if parsed.Error {
			return nil, fmt.Errorf("open-meteo: %s", parsed.Reason)
		}
		if err := assertUnits(parsed.HourlyUnits); err != nil {
			return nil, err
		}
		return mapHours(parsed)
	}
	return nil, last
}

func assertUnits(u map[string]string) error {
	if u == nil {
		return fmt.Errorf("open-meteo missing hourly_units")
	}
	if u["wind_speed_250hPa"] != "km/h" {
		return fmt.Errorf("unexpected wind unit %q (want km/h)", u["wind_speed_250hPa"])
	}
	if u["visibility"] != "m" && u["visibility"] != "" {
		return fmt.Errorf("unexpected visibility unit %q", u["visibility"])
	}
	return nil
}

func mapHours(p omResp) ([]domain.WeatherHour, error) {
	n := len(p.Hourly.Time)
	if n == 0 {
		return nil, fmt.Errorf("open-meteo empty hourly.time")
	}
	fields := []int{
		len(p.Hourly.CloudLow), len(p.Hourly.CloudMid), len(p.Hourly.CloudHigh),
		len(p.Hourly.RH), len(p.Hourly.Temp), len(p.Hourly.Dew),
		len(p.Hourly.Vis), len(p.Hourly.Precip),
		len(p.Hourly.Wind10), len(p.Hourly.Gust),
		len(p.Hourly.W250), len(p.Hourly.W500), len(p.Hourly.W850),
	}
	for _, L := range fields {
		if L != n {
			return nil, fmt.Errorf("open-meteo array length mismatch")
		}
	}
	out := make([]domain.WeatherHour, 0, n)
	var prev time.Time
	for i := 0; i < n; i++ {
		t, err := time.ParseInLocation("2006-01-02T15:04", p.Hourly.Time[i], time.UTC)
		if err != nil {
			t, err = time.Parse(time.RFC3339, p.Hourly.Time[i])
			if err != nil {
				return nil, fmt.Errorf("open-meteo time: %w", err)
			}
		}
		if !prev.IsZero() && !t.After(prev) {
			return nil, fmt.Errorf("open-meteo time axis not monotonic")
		}
		prev = t
		out = append(out, domain.WeatherHour{
			TimeUTC:     t,
			CloudLow:    clamp100(p.Hourly.CloudLow[i]),
			CloudMid:    clamp100(p.Hourly.CloudMid[i]),
			CloudHigh:   clamp100(p.Hourly.CloudHigh[i]),
			RH:          clamp100(p.Hourly.RH[i]),
			TempC:       p.Hourly.Temp[i],
			DewC:        p.Hourly.Dew[i],
			VisibilityM: p.Hourly.Vis[i],
			PrecipProb:  clamp100(p.Hourly.Precip[i]),
			Wind10MS:    KmhToMS(p.Hourly.Wind10[i]),
			Gust10MS:    KmhToMS(p.Hourly.Gust[i]),
			Wind250MS:   KmhToMS(p.Hourly.W250[i]),
			Wind500MS:   KmhToMS(p.Hourly.W500[i]),
			Wind850MS:   KmhToMS(p.Hourly.W850[i]),
		})
	}
	return out, nil
}

func clamp100(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
