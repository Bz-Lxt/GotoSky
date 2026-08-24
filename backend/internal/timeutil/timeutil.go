package timeutil

import (
	"time"
)

// Beijing is GMT+8. Used for display defaults and container TZ alignment.
var Beijing = time.FixedZone("CST", 8*3600)

func NowBeijing() time.Time {
	return time.Now().In(Beijing)
}

func InSite(t time.Time, loc *time.Location) time.Time {
	if loc == nil {
		loc = Beijing
	}
	return t.In(loc)
}

// CivilDate returns the civil Y-M-D in the given location.
// NEVER use UTC Year/Month/Day for "tonight" — 00:00–07:59 would flip the day.
func CivilDate(t time.Time, loc *time.Location) (y int, m time.Month, d int) {
	return InSite(t, loc).Date()
}

func StartOfCivilDay(t time.Time, loc *time.Location) time.Time {
	y, m, d := CivilDate(t, loc)
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

func ParseLocation(name string) *time.Location {
	if name == "" {
		return Beijing
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return Beijing
	}
	return loc
}
