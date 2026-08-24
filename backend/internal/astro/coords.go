package astro

import (
	"math"
	"time"
)

type Horizon struct {
	Alt float64 `json:"alt"`
	Az  float64 `json:"az"`
}

// AltAz converts J2000 RA/Dec to topocentric Alt/Az (no refraction).
func AltAz(t time.Time, lat, lon, raHours, decDeg float64) Horizon {
	jd := JulianDate(t)
	lst := LocalSiderealDeg(jd, lon)
	ha := wrap180(lst - raHours*15)
	alt := asind(sind(lat)*sind(decDeg) + cosd(lat)*cosd(decDeg)*cosd(ha))
	az := wrap360(atan2d(-cosd(decDeg)*sind(ha), sind(decDeg)*cosd(lat)-cosd(decDeg)*sind(lat)*cosd(ha)))
	return Horizon{Alt: alt, Az: az}
}

// KastenYoung airmass. X→∞ as alt→0; we clamp alt to 0.1°.
func Airmass(altDeg float64) float64 {
	if altDeg < 0.1 {
		altDeg = 0.1
	}
	return 1 / (sind(altDeg) + 0.50572*math.Pow(altDeg+6.07995, -1.6364))
}

func HorizonMasked(az, alt float64, mask []struct{ Az, Alt float64 }) bool {
	if len(mask) == 0 {
		return false
	}
	// Linear interpolate mask altitude at az.
	n := len(mask)
	for i := 0; i < n; i++ {
		a := mask[i]
		b := mask[(i+1)%n]
		az1, az2 := a.Az, b.Az
		span := wrap360(az2 - az1)
		off := wrap360(az - az1)
		if off <= span || i == n-1 {
			f := 0.0
			if span > 0 {
				f = off / span
			}
			lim := a.Alt + f*(b.Alt-a.Alt)
			return alt < lim
		}
	}
	return false
}
