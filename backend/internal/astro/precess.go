package astro

// ApplyPrecession is a first-order J2000 → date correction in degrees.
// Deep-sky Alt/Az budget is 0.1°; for nearby decades this term is small.
func ApplyPrecession(raHours, decDeg float64, year float64) (float64, float64) {
	T := (year - 2000.0) / 100
	m := 1.2812323*T + 0.0003879*T*T
	n := 0.5567530*T - 0.0001185*T*T
	ra := raHours*15 + (m + n*sind(raHours*15)*tand(decDeg))
	dec := decDeg + n*cosd(raHours*15)
	return wrap360(ra) / 15, dec
}
