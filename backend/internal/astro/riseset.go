package astro

import "time"

type RiseSet struct {
	Rise    time.Time `json:"rise"`
	Transit time.Time `json:"transit"`
	Set     time.Time `json:"set"`
	AlwaysUp   bool `json:"always_up"`
	AlwaysDown bool `json:"always_down"`
}

// ObjectRiseSet finds rise / transit / set of a J2000 target on the civil day.
func ObjectRiseSet(day time.Time, lat, lon, raHours, decDeg, minAlt float64, loc *time.Location) RiseSet {
	if loc == nil {
		loc = time.UTC
	}
	y, m, d := day.In(loc).Date()
	start := time.Date(y, m, d, 0, 0, 0, 0, loc)
	var out RiseSet
	var maxAlt = -90.0
	var maxT time.Time
	var prevAlt float64
	var havePrev bool
	step := 3 * time.Minute
	for i := 0; i <= 24*20; i++ {
		t := start.Add(time.Duration(i) * step)
		alt := AltAz(t, lat, lon, raHours, decDeg).Alt
		if alt > maxAlt {
			maxAlt, maxT = alt, t
		}
		if havePrev {
			if prevAlt < minAlt && alt >= minAlt && out.Rise.IsZero() {
				out.Rise = interpolateCross(t.Add(-step), prevAlt, t, alt, minAlt)
			}
			if prevAlt >= minAlt && alt < minAlt && out.Set.IsZero() {
				out.Set = interpolateCross(t.Add(-step), prevAlt, t, alt, minAlt)
			}
		}
		prevAlt, havePrev = alt, true
	}
	out.Transit = maxT.UTC()
	if maxAlt < minAlt {
		out.AlwaysDown = true
	}
	if maxAlt > minAlt {
		low := 90.0
		for i := 0; i <= 24*20; i++ {
			t := start.Add(time.Duration(i) * step)
			alt := AltAz(t, lat, lon, raHours, decDeg).Alt
			if alt < low {
				low = alt
			}
		}
		if low >= minAlt {
			out.AlwaysUp = true
		}
	}
	return out
}

func interpolateCross(t0 time.Time, a0 float64, t1 time.Time, a1, target float64) time.Time {
	if a1 == a0 {
		return t1.UTC()
	}
	f := (target - a0) / (a1 - a0)
	return t0.Add(time.Duration(float64(t1.Sub(t0)) * f)).UTC()
}
