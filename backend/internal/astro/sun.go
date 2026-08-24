package astro

import "time"

type Equatorial struct {
	RAHours float64
	DecDeg  float64
}

// SunEquatorial implements Meeus ch.25 low-precision (error ≪ 0.01°).
func SunEquatorial(t time.Time) Equatorial {
	jd := JulianDate(t)
	n := jd - J2000
	L := wrap360(280.460 + 0.9856474*n)
	g := wrap360(357.528 + 0.9856003*n)
	lambda := L + 1.915*sind(g) + 0.020*sind(2*g)
	eps := 23.439 - 0.0000004*n
	ra := atan2d(cosd(eps)*sind(lambda), cosd(lambda))
	dec := asind(sind(eps) * sind(lambda))
	return Equatorial{RAHours: wrap360(ra) / 15, DecDeg: dec}
}

func SunAltitude(t time.Time, lat, lon float64) float64 {
	eq := SunEquatorial(t)
	return AltAz(t, lat, lon, eq.RAHours, eq.DecDeg).Alt
}
