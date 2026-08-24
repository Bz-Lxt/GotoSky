package engine

import "github.com/google/uuid"

const EngineVersion = "sve-1.0.0"

// Profile holds configurable weights and EMPIRICAL seeing coefficients.
type Profile struct {
	ID uuid.UUID
	WC, WS, WM, WA, WT, WL, WN float64
	SeeingBias     float64
	SeeingV250     float64
	SeeingShear    float64
	SeeingV10      float64
	SeeingBest     float64
	SeeingWorst    float64
}

func DefaultProfile() Profile {
	return Profile{
		ID:          uuid.MustParse("00000000-0000-4000-8000-000000000001"),
		WC: 0.30, WS: 0.20, WM: 0.15, WA: 0.13, WT: 0.12, WL: 0.06, WN: 0.04,
		SeeingBias: 0.65, SeeingV250: 0.028, SeeingShear: 0.020, SeeingV10: 0.045,
		SeeingBest: 0.8, SeeingWorst: 3.5,
	}
}

type Input struct {
	CloudLow, CloudMid, CloudHigh float64
	RH, TempC, DewC               float64
	VisibilityM                   float64
	PrecipProb                    float64
	Wind10MS, Gust10MS            float64
	Wind250MS, Wind500MS, Wind850MS float64
	SunAlt, MoonAlt               float64
	MoonK                         float64
	MoonSepDeg                    float64
	TargetAlt                     float64
	Airmass                       float64
	SQM                           float64
	MinAltitude                   float64
	HorizonBlocked                bool
	Narrowband                    bool
}

type Result struct {
	Score          int
	Tier           string
	C, S, M, A, T, L, N float64
	SeeingArcsec   float64
	SeeingDerived  bool
	GateReason     string
	LimitingFactor string
	EngineVersion  string
}
