package astro

import "time"

type Twilight struct {
	Sunset     time.Time `json:"sunset"`
	Civil      time.Time `json:"civil"`
	Nautical   time.Time `json:"nautical"`
	Astronomical time.Time `json:"astronomical"`
	AstroDawn  time.Time `json:"astro_dawn"`
	NautDawn   time.Time `json:"naut_dawn"`
	CivilDawn  time.Time `json:"civil_dawn"`
	Sunrise    time.Time `json:"sunrise"`
}

// NightEvents finds sun-altitude crossings on the civil day of t in loc.
func NightEvents(t time.Time, lat, lon float64, loc *time.Location) Twilight {
	if loc == nil {
		loc = time.UTC
	}
	y, m, d := t.In(loc).Date()
	start := time.Date(y, m, d, 12, 0, 0, 0, loc) // local noon
	var out Twilight
	out.Sunset = findCrossing(start, lat, lon, -0.833, true)
	out.Civil = findCrossing(start, lat, lon, -6, true)
	out.Nautical = findCrossing(start, lat, lon, -12, true)
	out.Astronomical = findCrossing(start, lat, lon, -18, true)
	out.AstroDawn = findCrossing(start.Add(12*time.Hour), lat, lon, -18, false)
	out.NautDawn = findCrossing(start.Add(12*time.Hour), lat, lon, -12, false)
	out.CivilDawn = findCrossing(start.Add(12*time.Hour), lat, lon, -6, false)
	out.Sunrise = findCrossing(start.Add(12*time.Hour), lat, lon, -0.833, false)
	return out
}

func findCrossing(from time.Time, lat, lon, target float64, evening bool) time.Time {
	// Sample every 2 min for 14h, then bisection.
	step := 2 * time.Minute
	prevT := from
	prevA := SunAltitude(from, lat, lon)
	for i := 0; i < 420; i++ {
		t := from.Add(time.Duration(i+1) * step)
		a := SunAltitude(t, lat, lon)
		crossed := false
		if evening {
			crossed = prevA >= target && a < target
		} else {
			crossed = prevA <= target && a > target
		}
		if crossed {
			lo, hi := prevT, t
			for k := 0; k < 18; k++ {
				mid := lo.Add(hi.Sub(lo) / 2)
				if evening {
					if SunAltitude(mid, lat, lon) >= target {
						lo = mid
					} else {
						hi = mid
					}
				} else {
					if SunAltitude(mid, lat, lon) <= target {
						lo = mid
					} else {
						hi = mid
					}
				}
			}
			return hi.UTC()
		}
		prevT, prevA = t, a
	}
	return time.Time{}
}
