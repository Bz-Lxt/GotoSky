package astro

import "math"

// ApparentAltitude applies Bennett's refraction formula (arcminutes → degrees).
func ApparentAltitude(trueAlt float64) float64 {
	if trueAlt < -1 {
		return trueAlt
	}
	R := 1.02 / tand(trueAlt+10.3/(trueAlt+5.11))
	return trueAlt + R/60
}

func TrueAltitude(appAlt float64) float64 {
	if appAlt < -1 {
		return appAlt
	}
	R := 1.0 / tand(appAlt+7.31/(appAlt+4.4))
	return appAlt - R/60
}

func HorizontalRefractionArcmin(tempC, pressureHPa float64) float64 {
	if pressureHPa <= 0 {
		pressureHPa = 1013.25
	}
	return 1.02 * (pressureHPa / 1010) * (283 / (273 + tempC))
}

func ExtinctionMag(airmass, k float64) float64 {
	if k <= 0 {
		k = 0.20
	}
	return k * math.Max(airmass, 1)
}
