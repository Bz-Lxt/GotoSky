package engine

import (
	"fmt"
	"sort"

	"github.com/gotosky/gotosky/internal/domain"
)

type Candidate struct {
	Target    domain.Target `json:"target"`
	MeanScore float64       `json:"mean_score"`
	PeakScore int           `json:"peak_score"`
	FOVFit    float64       `json:"fov_fit"`
	Reason    string        `json:"reason"`
}

// Recommend picks top 1–3 explainable targets. Not a random draw (C8).
func Recommend(cands []Candidate, maxN int) []Candidate {
	if maxN <= 0 {
		maxN = 3
	}
	type scored struct {
		c Candidate
		v float64
	}
	ss := make([]scored, 0, len(cands))
	for _, c := range cands {
		if c.MeanScore < 50 {
			continue
		}
		fit := c.FOVFit
		if fit <= 0 {
			fit = 0.7
		}
		ss = append(ss, scored{c: c, v: c.MeanScore * fit})
	}
	sort.Slice(ss, func(i, j int) bool { return ss[i].v > ss[j].v })
	if len(ss) > maxN {
		ss = ss[:maxN]
	}
	out := make([]Candidate, 0, len(ss))
	for _, x := range ss {
		c := x.c
		if c.Reason == "" {
			c.Reason = fmt.Sprintf("当夜均分 %.0f、峰值 %d，视场匹配 %.0f%%，适合当前 Rig", c.MeanScore, c.PeakScore, c.FOVFit*100)
		}
		out = append(out, c)
	}
	return out
}

// FOVFit 1.0 when target angular size sits comfortably in the frame (20–80% of FOV).
func FOVFit(targetArcmin, fovArcmin float64) float64 {
	if fovArcmin <= 0 || targetArcmin <= 0 {
		return 0.7
	}
	ratio := targetArcmin / fovArcmin
	switch {
	case ratio < 0.05:
		return 0.35
	case ratio > 1.2:
		return 0.25
	case ratio >= 0.2 && ratio <= 0.8:
		return 1.0
	case ratio < 0.2:
		return 0.35 + (ratio-0.05)/0.15*0.65
	default:
		return 1.0 - (ratio-0.8)/0.4*0.75
	}
}
