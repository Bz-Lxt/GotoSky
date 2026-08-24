package weather

import "testing"

func TestKmhToMS(t *testing.T) {
	if got := ConvertKmh(77.5); got < 21.5 || got > 21.6 {
		t.Fatalf("77.5 km/h → %v m/s want ~21.53", got)
	}
	if KmhToMS(3.6) != 1 {
		t.Fatal(KmhToMS(3.6))
	}
}

func TestMockDeterministic(t *testing.T) {
	m := NewMock("CLEAR")
	a, _ := m.Forecast(nil, 40.45, 116.02, 2)
	b, _ := m.Forecast(nil, 40.45, 116.02, 2)
	if len(a) != 48 || a[0].Wind250MS != b[0].Wind250MS {
		t.Fatalf("len=%d", len(a))
	}
	if a[0].Wind250MS < 20 || a[0].Wind250MS > 23 {
		t.Fatalf("expected ~21.5 m/s already SI, got %v", a[0].Wind250MS)
	}
}
