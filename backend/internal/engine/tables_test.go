package engine

import "testing"

func TestFactorMonotonicCloud(t *testing.T) {
	p := DefaultProfile()
	base := Input{
		RH: 40, TempC: 10, DewC: 2, VisibilityM: 25000, PrecipProb: 0,
		Wind10MS: 2, Gust10MS: 3, Wind250MS: 15, Wind500MS: 10, Wind850MS: 3,
		SunAlt: -20, MoonAlt: -5, MoonK: 0.1, MoonSepDeg: 90,
		TargetAlt: 60, Airmass: 1.15, SQM: 21.5, MinAltitude: 20,
	}
	var prev int = 101
	for _, cl := range []float64{0, 10, 20, 35, 50} {
		in := base
		in.CloudLow, in.CloudMid, in.CloudHigh = cl, cl*0.6, cl*0.3
		r := Evaluate(in, p)
		if r.Score > prev {
			t.Fatalf("cloud %v score rose %d > %d", cl, r.Score, prev)
		}
		prev = r.Score
	}
}

func TestNarrowbandReducesMoonPenalty(t *testing.T) {
	p := DefaultProfile()
	in := Input{
		CloudLow: 5, CloudMid: 5, CloudHigh: 5, RH: 40, TempC: 8, DewC: 1, VisibilityM: 20000,
		Wind10MS: 2, Gust10MS: 3, Wind250MS: 12, Wind500MS: 8, Wind850MS: 2,
		SunAlt: -20, MoonAlt: 45, MoonK: 0.95, MoonSepDeg: 20,
		TargetAlt: 55, Airmass: 1.2, SQM: 21.0, MinAltitude: 20,
	}
	wide := Evaluate(in, p)
	in.Narrowband = true
	nb := Evaluate(in, p)
	if nb.M <= wide.M {
		t.Fatalf("narrowband should lift M: wide=%v nb=%v", wide.M, nb.M)
	}
}

func TestAllGates(t *testing.T) {
	p := DefaultProfile()
	good := Input{
		CloudLow: 5, VisibilityM: 20000, SunAlt: -20, TargetAlt: 50, MinAltitude: 20, SQM: 21, Gust10MS: 2, RH: 40, TempC: 8, DewC: 1,
	}
	cases := []struct {
		mut  func(*Input)
		want string
	}{
		{func(in *Input) { in.CloudLow, in.CloudMid, in.CloudHigh = 90, 80, 70 }, "G_CLOUD"},
		{func(in *Input) { in.PrecipProb = 55 }, "G_RAIN"},
		{func(in *Input) { in.Gust10MS = 12 }, "G_WIND"},
		{func(in *Input) { in.SunAlt = -8 }, "G_NIGHT"},
		{func(in *Input) { in.TargetAlt = 10 }, "G_ALT"},
		{func(in *Input) { in.HorizonBlocked = true }, "G_ALT"},
	}
	for _, c := range cases {
		in := good
		c.mut(&in)
		r := Evaluate(in, p)
		if r.GateReason != c.want || r.Score != 0 {
			t.Fatalf("want %s got %+v", c.want, r)
		}
	}
}

func TestBortleMap(t *testing.T) {
	if BortleToSQM(1) != 21.9 || BortleToSQM(9) != 17.6 {
		t.Fatal(BortleToSQM(1), BortleToSQM(9))
	}
}
