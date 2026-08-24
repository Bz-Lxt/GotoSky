package astro

// BrightStar is a background star for the virtual sky canvas.
type BrightStar struct {
	Name    string  `json:"name"`
	RAHours float64 `json:"ra_hours"`
	DecDeg  float64 `json:"dec_deg"`
	Mag     float64 `json:"mag"`
}

func BrightStars() []BrightStar {
	return []BrightStar{
		{"Sirius", 6.752, -16.72, -1.46},
		{"Canopus", 6.399, -52.70, -0.74},
		{"Arcturus", 14.261, 19.18, -0.05},
		{"Vega", 18.615, 38.78, 0.03},
		{"Capella", 5.278, 45.99, 0.08},
		{"Rigel", 5.242, -8.20, 0.13},
		{"Procyon", 7.655, 5.23, 0.34},
		{"Betelgeuse", 5.919, 7.41, 0.50},
		{"Altair", 19.846, 8.87, 0.77},
		{"Aldebaran", 4.598, 16.51, 0.85},
		{"Antares", 16.490, -26.43, 0.96},
		{"Spica", 13.420, -11.16, 0.98},
		{"Pollux", 7.755, 28.03, 1.14},
		{"Fomalhaut", 22.960, -29.62, 1.16},
		{"Deneb", 20.690, 45.28, 1.25},
		{"Regulus", 10.139, 11.97, 1.35},
		{"Castor", 7.576, 31.89, 1.57},
		{"Polaris", 2.530, 89.26, 1.98},
		{"Dubhe", 11.062, 61.75, 1.79},
		{"Alioth", 12.900, 55.96, 1.77},
	}
}
