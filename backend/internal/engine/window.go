package engine

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/gotosky/gotosky/internal/domain"
	"github.com/gotosky/gotosky/internal/timeutil"
)

type SlotScore struct {
	At    time.Time
	Score int
	Tier  string
	Limit string
}

// Windows merges Score≥50 adjacent hours, drops <60min, ranks by QI.
func Windows(siteID, targetID, profileID uuid.UUID, loc *time.Location, slots []SlotScore) []domain.GoldenWindow {
	if loc == nil {
		loc = timeutil.Beijing
	}
	var runs [][]SlotScore
	var cur []SlotScore
	for _, s := range slots {
		if s.Score >= 50 {
			if len(cur) > 0 && s.At.Sub(cur[len(cur)-1].At) > time.Hour+time.Minute {
				runs = append(runs, cur)
				cur = nil
			}
			cur = append(cur, s)
		} else if len(cur) > 0 {
			runs = append(runs, cur)
			cur = nil
		}
	}
	if len(cur) > 0 {
		runs = append(runs, cur)
	}
	out := make([]domain.GoldenWindow, 0)
	for _, run := range runs {
		if len(run) < 1 {
			continue
		}
		dur := run[len(run)-1].At.Sub(run[0].At) + time.Hour
		if dur < time.Hour {
			continue
		}
		sum := 0
		peak := 0
		lim := run[0].Limit
		for _, s := range run {
			sum += s.Score
			if s.Score > peak {
				peak = s.Score
				lim = s.Limit
			}
		}
		mean := float64(sum) / float64(len(run))
		qi := mean * dur.Hours()
		w := domain.GoldenWindow{
			ID:              uuid.New(),
			SiteID:          siteID,
			TargetID:        targetID,
			StartUTC:        run[0].At.UTC(),
			EndUTC:          run[len(run)-1].At.Add(time.Hour).UTC(),
			StartLocal:      run[0].At.In(loc),
			EndLocal:        run[len(run)-1].At.Add(time.Hour).In(loc),
			Tier:            tierOf(int(mean + 0.5)),
			MeanScore:       mean,
			PeakScore:       peak,
			QualityIntegral: qi,
			LimitingFactor:  lim,
			EngineVersion:   EngineVersion,
			WeightProfileID: profileID,
		}
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].QualityIntegral > out[j].QualityIntegral })
	if len(out) > 3 {
		out = out[:3]
	}
	return out
}
