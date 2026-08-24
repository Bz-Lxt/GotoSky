package handler

func validLatLon(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

func validBortle(b int) bool { return b >= 1 && b <= 9 }

func validExposure(s float64) bool { return s > 0 && s <= 3600 }
