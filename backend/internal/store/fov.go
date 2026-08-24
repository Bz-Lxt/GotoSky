package store

import "math"

// PlateScaleArcsec returns arcsec/pixel = 206.265 * pix_um / focal_mm.
func PlateScaleArcsec(pixUM, focalMM float64) float64 {
	if focalMM <= 0 {
		return 0
	}
	return 206.265 * pixUM / focalMM
}

func FOVArcmin(pixUM, widthPx, focalMM float64) float64 {
	s := PlateScaleArcsec(pixUM, focalMM)
	return s * widthPx / 60
}

func SamplingOK(scale, fwhmArcsec float64) bool {
	if scale <= 0 {
		return false
	}
	return fwhmArcsec/scale >= 2 && fwhmArcsec/scale <= 4 || math.Abs(fwhmArcsec/scale-2.5) < 2
}
