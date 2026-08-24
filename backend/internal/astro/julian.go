package astro

import "time"

// JulianDate returns JD for a UTC instant (TT≈UTC for our precision budget).
func JulianDate(t time.Time) float64 {
	t = t.UTC()
	y := t.Year()
	m := int(t.Month())
	d := float64(t.Day()) + (float64(t.Hour())+float64(t.Minute())/60+float64(t.Second())/3600+float64(t.Nanosecond())/1e9/3600)/24
	if m <= 2 {
		y--
		m += 12
	}
	A := y / 100
	B := 2 - A + A/4
	return float64(int(365.25*float64(y+4716))) + float64(int(30.6001*float64(m+1))) + d + float64(B) - 1524.5
}

func JulianCentury(jd float64) float64 {
	return (jd - J2000) / 36525
}

// GMSTDeg is Greenwich mean sidereal time in degrees (Meeus 12.4).
func GMSTDeg(jd float64) float64 {
	T := JulianCentury(jd)
	theta := 280.46061837 + 360.98564736629*(jd-J2000) + 0.000387933*T*T - T*T*T/38710000
	return wrap360(theta)
}

func LocalSiderealDeg(jd, lonEast float64) float64 {
	return wrap360(GMSTDeg(jd) + lonEast)
}
