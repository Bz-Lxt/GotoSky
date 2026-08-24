package engine

import (
	"math"
	"strings"
)

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func pow(x, p float64) float64 {
	if x <= 0 {
		return 0
	}
	return math.Pow(x, p)
}

func factorFloor(v float64) float64 { return clamp(v, 0.02, 1) }

// Evaluate is a pure function: weather + sky geometry → 0–100 score.
func Evaluate(in Input, p Profile) Result {
	r := Result{SeeingDerived: true, EngineVersion: EngineVersion}
	if g, why := hardGates(in); g == 0 {
		r.Score = 0
		r.Tier = "UNUSABLE"
		r.GateReason = why
		r.C, r.S, r.M, r.A, r.T, r.L, r.N = 0.02, 0.02, 0.02, 0.02, 0.02, 0.02, 0.02
		r.SeeingArcsec = seeingEst(in, p)
		r.LimitingFactor = why
		return r
	}
	r.C = cloudFactor(in)
	r.S, r.SeeingArcsec = seeingFactor(in, p)
	r.M = moonFactor(in)
	r.A = altitudeFactor(in)
	r.T = transparencyFactor(in)
	r.L = lightFactor(in)
	r.N = nightFactor(in)

	q := pow(r.C, p.WC) * pow(r.S, p.WS) * pow(r.M, p.WM) * pow(r.A, p.WA) *
		pow(r.T, p.WT) * pow(r.L, p.WL) * pow(r.N, p.WN)
	r.Score = int(math.Round(100 * q))
	if r.Score < 0 {
		r.Score = 0
	}
	if r.Score > 100 {
		r.Score = 100
	}
	r.Tier = tierOf(r.Score)
	r.LimitingFactor = limiting(r, p)
	return r
}

func hardGates(in Input) (float64, string) {
	cloudEff := 0.50*in.CloudLow + 0.30*in.CloudMid + 0.20*in.CloudHigh
	if cloudEff > 60 {
		return 0, "G_CLOUD"
	}
	if in.PrecipProb > 30 {
		return 0, "G_RAIN"
	}
	if in.Gust10MS > 11.11 { // 40 km/h
		return 0, "G_WIND"
	}
	if in.SunAlt > -12 {
		return 0, "G_NIGHT"
	}
	if in.TargetAlt < in.MinAltitude || in.HorizonBlocked {
		return 0, "G_ALT"
	}
	return 1, ""
}

func cloudFactor(in Input) float64 {
	cloudEff := 0.50*in.CloudLow + 0.30*in.CloudMid + 0.20*in.CloudHigh
	return factorFloor(pow(1-cloudEff/100, 1.6))
}

func seeingEst(in Input, p Profile) float64 {
	shear := math.Abs(in.Wind500MS - in.Wind850MS)
	s := p.SeeingBias + p.SeeingV250*in.Wind250MS + p.SeeingShear*shear + p.SeeingV10*in.Wind10MS
	return clamp(s, 0.5, 6.0)
}

func seeingFactor(in Input, p Profile) (float64, float64) {
	est := seeingEst(in, p)
	s := clamp((p.SeeingWorst-est)/(p.SeeingWorst-p.SeeingBest), 0.02, 1)
	return s, est
}

func moonFactor(in Input) float64 {
	if in.MoonAlt < -0.8 {
		return 1
	}
	altF := pow(clamp(math.Sin(in.MoonAlt*math.Pi/180), 0, 1), 0.7)
	sepF := clamp(1-in.MoonSepDeg/110, 0.05, 1)
	m := clamp(1-pow(in.MoonK, 1.1)*altF*sepF, 0.02, 1)
	if in.Narrowband {
		m = 1 - 0.65*(1-m)
	}
	return factorFloor(m)
}

func altitudeFactor(in Input) float64 {
	x := in.Airmass
	if x <= 0 {
		x = 1
	}
	return factorFloor(pow(clamp((2.6-x)/(2.6-1.0), 0.02, 1), 0.85))
}

func transparencyFactor(in Input) float64 {
	rhF := clamp((96-in.RH)/46, 0.02, 1)
	dewF := clamp((in.TempC-in.DewC)/4.0, 0.15, 1)
	visF := clamp(in.VisibilityM/20000, 0.30, 1)
	return factorFloor(pow(rhF, 0.50) * pow(dewF, 0.35) * pow(visF, 0.15))
}

func DewRisk(in Input) bool {
	return in.RH >= 95 && (in.TempC-in.DewC) <= 1.0
}

func lightFactor(in Input) float64 {
	return factorFloor(clamp((in.SQM-17.8)/(21.9-17.8), 0.05, 1))
}

func nightFactor(in Input) float64 {
	n := clamp((-in.SunAlt-12)/6, 0, 1)*0.7 + 0.3
	return factorFloor(n)
}

func tierOf(score int) string {
	switch {
	case score >= 80:
		return "GOLD"
	case score >= 65:
		return "SILVER"
	case score >= 50:
		return "BRONZE"
	case score >= 1:
		return "POOR"
	default:
		return "UNUSABLE"
	}
}

func limiting(r Result, p Profile) string {
	type pair struct {
		name string
		loss float64
	}
	fs := []pair{
		{"CLOUD", p.WC * -math.Log(math.Max(r.C, 1e-9))},
		{"SEEING", p.WS * -math.Log(math.Max(r.S, 1e-9))},
		{"MOON", p.WM * -math.Log(math.Max(r.M, 1e-9))},
		{"ALTITUDE", p.WA * -math.Log(math.Max(r.A, 1e-9))},
		{"TRANSPARENCY", p.WT * -math.Log(math.Max(r.T, 1e-9))},
		{"LIGHT_POLLUTION", p.WL * -math.Log(math.Max(r.L, 1e-9))},
		{"NIGHT", p.WN * -math.Log(math.Max(r.N, 1e-9))},
	}
	best := fs[0]
	for _, x := range fs[1:] {
		if x.loss > best.loss {
			best = x
		}
	}
	return best.name
}

func BortleToSQM(b int) float64 {
	m := map[int]float64{1: 21.9, 2: 21.7, 3: 21.5, 4: 21.0, 5: 20.4, 6: 19.4, 7: 18.6, 8: 18.0, 9: 17.6}
	if v, ok := m[b]; ok {
		return v
	}
	return 20.4
}

func FactorLabel(s string) string { return strings.ToUpper(s) }
