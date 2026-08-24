package astro

import "math"

const (
	Deg2Rad = math.Pi / 180
	Rad2Deg = 180 / math.Pi
	J2000   = 2451545.0
)

func wrap360(d float64) float64 {
	d = math.Mod(d, 360)
	if d < 0 {
		d += 360
	}
	return d
}

func wrap180(d float64) float64 {
	d = wrap360(d)
	if d > 180 {
		d -= 360
	}
	return d
}

func sind(d float64) float64 { return math.Sin(d * Deg2Rad) }
func cosd(d float64) float64 { return math.Cos(d * Deg2Rad) }
func tand(d float64) float64 { return math.Tan(d * Deg2Rad) }

func asind(x float64) float64 { return math.Asin(clamp(x, -1, 1)) * Rad2Deg }
func acosd(x float64) float64 { return math.Acos(clamp(x, -1, 1)) * Rad2Deg }
func atan2d(y, x float64) float64 {
	return math.Atan2(y, x) * Rad2Deg
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
