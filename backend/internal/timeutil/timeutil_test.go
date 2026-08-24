package timeutil

import (
	"testing"
	"time"
)

func TestCivilDateNotUTC(t *testing.T) {
	// 2026-08-24 02:30 UTC = 10:30 Beijing same day
	// 2026-08-24 18:30 UTC = 2026-08-25 02:30 Beijing — must flip
	u := time.Date(2026, 8, 24, 18, 30, 0, 0, time.UTC)
	y, m, d := CivilDate(u, Beijing)
	if y != 2026 || m != 8 || d != 25 {
		t.Fatalf("got %d-%d-%d", y, m, d)
	}
	uy, um, ud := u.Date()
	if ud == d && um == m && uy == y {
		t.Fatal("civil date accidentally used UTC")
	}
}
