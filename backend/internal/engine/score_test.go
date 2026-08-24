package engine

import "testing"

func TestSeeingCalibration(t *testing.T) {
	p := DefaultProfile()
	in := Input{
		Wind250MS: 21.5, Wind500MS: 12.5, Wind850MS: 2.4, Wind10MS: 2.2,
		CloudLow: 8, CloudMid: 6, CloudHigh: 4, RH: 40, TempC: 10, DewC: 2,
		VisibilityM: 25000, PrecipProb: 0, Gust10MS: 3,
		SunAlt: -20, MoonAlt: -10, MoonK: 0.2, MoonSepDeg: 80,
		TargetAlt: 55, Airmass: 1.22, SQM: 21.0, MinAltitude: 20,
	}
	est := seeingEst(in, p)
	if est < 1.4 || est > 1.7 {
		t.Fatalf("seeing_est=%v want ~1.55", est)
	}
	r := Evaluate(in, p)
	if r.Score < 50 {
		t.Fatalf("score %d too low for clear night", r.Score)
	}
	if r.Tier == "UNUSABLE" {
		t.Fatal(r.GateReason)
	}
	if !r.SeeingDerived {
		t.Fatal("seeing must be marked derived")
	}
}

func TestHardGates(t *testing.T) {
	p := DefaultProfile()
	in := Input{CloudLow: 90, CloudMid: 80, CloudHigh: 70, SunAlt: -20, TargetAlt: 50, MinAltitude: 20, VisibilityM: 10000, SQM: 21}
	r := Evaluate(in, p)
	if r.Score != 0 || r.GateReason != "G_CLOUD" {
		t.Fatalf("%+v", r)
	}
	in = Input{CloudLow: 10, SunAlt: -5, TargetAlt: 50, MinAltitude: 20, VisibilityM: 10000, SQM: 21, Gust10MS: 1}
	r = Evaluate(in, p)
	if r.GateReason != "G_NIGHT" {
		t.Fatalf("want G_NIGHT got %s", r.GateReason)
	}
}

func TestWeightsSum(t *testing.T) {
	p := DefaultProfile()
	s := p.WC + p.WS + p.WM + p.WA + p.WT + p.WL + p.WN
	if s < 0.999 || s > 1.001 {
		t.Fatalf("weights %v", s)
	}
}

func TestWindowsMerge(t *testing.T) {
	t0 := mustParse("2026-08-24T13:00:00Z")
	var slots []SlotScore
	for i := 0; i < 5; i++ {
		slots = append(slots, SlotScore{At: t0.Add(hours(i)), Score: 70, Limit: "SEEING"})
	}
	slots = append(slots, SlotScore{At: t0.Add(hours(6)), Score: 10})
	w := Windows(DefaultProfile().ID, DefaultProfile().ID, DefaultProfile().ID, nil, slots)
	if len(w) != 1 || w[0].PeakScore != 70 {
		t.Fatalf("%+v", w)
	}
}
