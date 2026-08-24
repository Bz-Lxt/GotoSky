package weather

// ConvertKmh is the single unit boundary for Open-Meteo km/h → m/s.
func ConvertKmh(kmh float64) float64 { return KmhToMS(kmh) }
