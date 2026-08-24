package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gotosky/gotosky/internal/astro"
	"github.com/gotosky/gotosky/internal/domain"
	"github.com/gotosky/gotosky/internal/engine"
	"github.com/gotosky/gotosky/internal/logger"
	"github.com/gotosky/gotosky/internal/store"
	"github.com/gotosky/gotosky/internal/timeutil"
	"github.com/gotosky/gotosky/internal/weather"
)

type Scorer struct {
	Store   *store.Store
	Weather weather.Provider
	Profile engine.Profile
	mu      sync.Mutex
	running map[uuid.UUID]bool
}

func NewScorer(st *store.Store, w weather.Provider) *Scorer {
	return &Scorer{Store: st, Weather: w, Profile: engine.DefaultProfile(), running: map[uuid.UUID]bool{}}
}

func (s *Scorer) Recompute(ctx context.Context, site domain.Site, targets []domain.Target, days int) error {
	s.mu.Lock()
	if s.running[site.ID] {
		s.mu.Unlock()
		return nil
	}
	s.running[site.ID] = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.running, site.ID)
		s.mu.Unlock()
	}()

	hours, err := s.Weather.Forecast(ctx, site.Latitude, site.Longitude, days)
	if err != nil {
		return err
	}
	if len(hours) > 0 {
		raw, _ := json.Marshal(hours)
		_ = s.Store.InsertWeather(ctx, site.ID, s.Weather.Name(), hours[0].TimeUTC, hours[len(hours)-1].TimeUTC, raw, false)
	}
	loc := timeutil.ParseLocation(site.Timezone)
	for _, tgt := range targets {
		if err := s.scoreOne(ctx, site, tgt, hours, loc); err != nil {
			logger.L().Error("score target", "target", tgt.CatalogID, "err", err)
		}
	}
	return nil
}

func (s *Scorer) scoreOne(ctx context.Context, site domain.Site, tgt domain.Target, hours []domain.WeatherHour, loc *time.Location) error {
	slots := make([]domain.ScoreSlot, 0, len(hours))
	winIn := make([]engine.SlotScore, 0, len(hours))
	for _, h := range hours {
		in := engine.Input{
			CloudLow: h.CloudLow, CloudMid: h.CloudMid, CloudHigh: h.CloudHigh,
			RH: h.RH, TempC: h.TempC, DewC: h.DewC, VisibilityM: h.VisibilityM, PrecipProb: h.PrecipProb,
			Wind10MS: h.Wind10MS, Gust10MS: h.Gust10MS, Wind250MS: h.Wind250MS, Wind500MS: h.Wind500MS, Wind850MS: h.Wind850MS,
			SunAlt: astro.SunAltitude(h.TimeUTC, site.Latitude, site.Longitude),
			MoonAlt: astro.MoonAltitude(h.TimeUTC, site.Latitude, site.Longitude),
			MoonK: astro.MoonIllumination(h.TimeUTC),
			MoonSepDeg: astro.MoonTargetSepDeg(h.TimeUTC, tgt.RAHours, tgt.DecDeg),
			SQM: site.SQM, MinAltitude: site.MinAltitude,
		}
		hz := astro.AltAz(h.TimeUTC, site.Latitude, site.Longitude, tgt.RAHours, tgt.DecDeg)
		in.TargetAlt = hz.Alt
		in.Airmass = astro.Airmass(hz.Alt)
		r := engine.Evaluate(in, s.Profile)
		sl := domain.ScoreSlot{
			ID: uuid.New(), SiteID: site.ID, TargetID: tgt.ID, SlotUTC: h.TimeUTC,
			Score: r.Score, Tier: r.Tier,
			FactorC: r.C, FactorS: r.S, FactorM: r.M, FactorA: r.A, FactorT: r.T, FactorL: r.L, FactorN: r.N,
			SeeingArcsec: r.SeeingArcsec, SeeingDerived: true,
			GateReason: r.GateReason, LimitingFactor: r.LimitingFactor,
			EngineVersion: engine.EngineVersion, WeightProfileID: s.Profile.ID,
		}
		slots = append(slots, sl)
		winIn = append(winIn, engine.SlotScore{At: h.TimeUTC, Score: r.Score, Tier: r.Tier, Limit: r.LimitingFactor})
		if engine.DewRisk(in) {
			sid := site.ID
			s.Store.InsertAlert(ctx, "DEW_RISK", "结露风险 "+h.TimeUTC.In(loc).Format("01-02 15:04"), &sid, nil)
		}
	}
	if err := s.Store.ReplaceScores(ctx, slots); err != nil {
		return err
	}
	wins := engine.Windows(site.ID, tgt.ID, s.Profile.ID, loc, winIn)
	return s.Store.ReplaceWindows(ctx, site.ID, tgt.ID, wins)
}

func (s *Scorer) SkyTrack(site domain.Site, tgt domain.Target, from time.Time, hours int) []map[string]any {
	out := make([]map[string]any, 0, hours)
	for i := 0; i < hours; i++ {
		t := from.Add(time.Duration(i) * time.Hour)
		hz := astro.AltAz(t, site.Latitude, site.Longitude, tgt.RAHours, tgt.DecDeg)
		moon := astro.MoonEquatorial(t)
		mhz := astro.AltAz(t, site.Latitude, site.Longitude, moon.RAHours, moon.DecDeg)
		out = append(out, map[string]any{
			"t": t.UTC(), "alt": hz.Alt, "az": hz.Az,
			"moon_alt": mhz.Alt, "moon_az": mhz.Az,
			"sun_alt": astro.SunAltitude(t, site.Latitude, site.Longitude),
			"airmass": astro.Airmass(hz.Alt),
		})
	}
	return out
}
