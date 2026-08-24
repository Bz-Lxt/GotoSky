package astro

import "time"

// MoonEquatorial uses Meeus ch.47 principal terms.
// Declared accuracy budget: RA/Dec ≤ 0.05°.
func MoonEquatorial(t time.Time) Equatorial {
	jd := JulianDate(t)
	T := JulianCentury(jd)
	Lp := wrap360(218.3164477 + 481267.88123421*T - 0.0015786*T*T + T*T*T/538841 - T*T*T*T/65194000)
	D := wrap360(297.8501921 + 445267.1114034*T - 0.0018819*T*T + T*T*T/545868 - T*T*T*T/113065000)
	M := wrap360(357.5291092 + 35999.0502909*T - 0.0001536*T*T + T*T*T/24490000)
	Mp := wrap360(134.9633964 + 477198.8675055*T + 0.0087414*T*T + T*T*T/69699 - T*T*T*T/14712000)
	F := wrap360(93.2720950 + 483202.0175233*T - 0.0036539*T*T - T*T*T/3526000 + T*T*T*T/863310000)
	E := 1 - 0.002516*T - 0.0000074*T*T

	// Longitude periodic terms (arcsec), principal set.
	sl := 6288774*sind(Mp) + 1274027*sind(2*D-Mp) + 658314*sind(2*D) + 213618*sind(2*Mp) -
		185116*E*sind(M) - 114332*sind(2*F) + 58793*sind(2*D-2*Mp) + 57066*E*sind(2*D-M-Mp) +
		53322*sind(2*D+Mp) + 45758*E*sind(2*D-M) - 40923*E*sind(M-Mp) - 34720*sind(D) -
		30383*E*sind(M+Mp) + 15327*sind(2*D-2*F) - 12528*sind(Mp+2*F) + 10980*sind(Mp-2*F) +
		10675*sind(4*D-Mp) + 10034*sind(3*Mp) + 8548*sind(4*D-2*Mp) - 7888*E*sind(2*D+M-Mp) -
		6766*E*sind(2*D+M) - 5163*sind(D-Mp) + 4987*E*sind(D+M) + 4036*E*sind(2*D-M+Mp)
	sb := 5128122*sind(F) + 280602*sind(Mp+F) + 277693*sind(Mp-F) + 173237*sind(2*D-F) +
		55413*sind(2*D-Mp+F) + 46271*sind(2*D-Mp-F) + 32573*sind(2*D+F) + 17198*sind(2*Mp+F) +
		9266*sind(2*D+Mp-F) + 8822*sind(2*Mp-F) + 8216*E*sind(2*D-M-F)

	lambda := wrap360(Lp + sl/1e6)
	beta := sb / 1e6
	eps := 23.439291 - 0.0130042*T
	ra := atan2d(sind(lambda)*cosd(eps)-tand(beta)*sind(eps), cosd(lambda))
	dec := asind(sind(beta)*cosd(eps) + cosd(beta)*sind(eps)*sind(lambda))
	return Equatorial{RAHours: wrap360(ra) / 15, DecDeg: dec}
}

func MoonAltitude(t time.Time, lat, lon float64) float64 {
	eq := MoonEquatorial(t)
	return AltAz(t, lat, lon, eq.RAHours, eq.DecDeg).Alt
}

// MoonIllumination k in [0,1] (Meeus 48).
func MoonIllumination(t time.Time) float64 {
	sun := SunEquatorial(t)
	moon := MoonEquatorial(t)
	psi := angularSepDeg(sun.RAHours*15, sun.DecDeg, moon.RAHours*15, moon.DecDeg)
	i := 180 - psi
	k := (1 + cosd(i)) / 2
	return clamp(k, 0, 1)
}

func angularSepDeg(ra1, dec1, ra2, dec2 float64) float64 {
	c := sind(dec1)*sind(dec2) + cosd(dec1)*cosd(dec2)*cosd(ra1-ra2)
	return acosd(c)
}

func MoonTargetSepDeg(t time.Time, raH, dec float64) float64 {
	m := MoonEquatorial(t)
	return angularSepDeg(m.RAHours*15, m.DecDeg, raH*15, dec)
}
