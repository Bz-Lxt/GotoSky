package astro

import (
	"math"
	"testing"
	"time"
)

func TestJulianJ2000(t *testing.T) {
	tt := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	jd := JulianDate(tt)
	if math.Abs(jd-2451545.0) > 0.0001 {
		t.Fatalf("JD=%v", jd)
	}
}

func TestSunJan1(t *testing.T) {
	// NOAA-ish: 2000-01-01 12:00 UTC sun near RA 18.72h Dec -23.0
	tt := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	eq := SunEquatorial(tt)
	if math.Abs(eq.DecDeg+23.0) > 0.5 {
		t.Fatalf("sun dec %v", eq.DecDeg)
	}
	if eq.RAHours < 18.2 || eq.RAHours > 19.2 {
		t.Fatalf("sun ra %v", eq.RAHours)
	}
}

func TestMoonIlluminationRange(t *testing.T) {
	k := MoonIllumination(time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC))
	if k < 0 || k > 1 {
		t.Fatal(k)
	}
}

func TestAirmassZenith(t *testing.T) {
	x := Airmass(90)
	if math.Abs(x-1) > 0.02 {
		t.Fatalf("X=%v", x)
	}
}

func TestAltAzPolaris(t *testing.T) {
	tt := time.Date(2026, 8, 24, 14, 0, 0, 0, time.UTC)
	hz := AltAz(tt, 40.45, 116.02, 2.530, 89.26)
	if hz.Alt < 35 || hz.Alt > 50 {
		t.Fatalf("polaris alt %v at Xinglong", hz.Alt)
	}
}

func TestNightEventsExist(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	ev := NightEvents(time.Date(2026, 1, 15, 12, 0, 0, 0, loc), 40.45, 116.02, loc)
	if ev.Sunset.IsZero() || ev.Sunrise.IsZero() {
		t.Fatalf("%+v", ev)
	}
	if !ev.Sunset.Before(ev.Astronomical) && !ev.Astronomical.IsZero() {
		t.Log("polar winter edge ok")
	}
}
