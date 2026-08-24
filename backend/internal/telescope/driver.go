package telescope

import (
	"context"
	"time"
)

type Sensors struct {
	MountRA      float64 `json:"mount_ra"`
	MountDec     float64 `json:"mount_dec"`
	Slewing      bool    `json:"slewing"`
	GuideRMS     float64 `json:"guide_rms"`
	CCDTemp      float64 `json:"ccd_temp"`
	FilterPos    int     `json:"filter_pos"`
	Humidity     float64 `json:"humidity"`
	WindMS       float64 `json:"wind_ms"`
	Source       string  `json:"source"`
}

type Driver interface {
	Name() string
	Source() string // SIMULATED | DEVICE
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Slew(ctx context.Context, raHours, decDeg float64) error
	WaitSlew(ctx context.Context) error
	Settle(ctx context.Context) error
	LockGuide(ctx context.Context) error
	SetFilter(ctx context.Context, pos int) error
	Expose(ctx context.Context, seconds float64) error
	Dither(ctx context.Context) error
	Park(ctx context.Context) error
	Heartbeat(ctx context.Context) error
	ReadSensors(ctx context.Context) (Sensors, error)
	Inject(fault string)
}

type FaultClass string

const (
	ClassTransient FaultClass = "TRANSIENT"
	ClassPermanent FaultClass = "PERMANENT"
	ClassSafety    FaultClass = "SAFETY"
)

type DeviceError struct {
	Code  string
	Class FaultClass
	Msg   string
}

func (e *DeviceError) Error() string { return e.Code + ": " + e.Msg }

func Classify(err error) FaultClass {
	if err == nil {
		return ""
	}
	if de, ok := err.(*DeviceError); ok {
		return de.Class
	}
	return ClassTransient
}

func TimeoutOf(state string, exposeSec float64) time.Duration {
	switch state {
	case "CONNECTING":
		return 10 * time.Second
	case "SLEWING":
		return 120 * time.Second
	case "SETTLING":
		return 30 * time.Second
	case "GUIDE_LOCKING":
		return 60 * time.Second
	case "FILTER_CHANGING":
		return 20 * time.Second
	case "EXPOSING":
		return time.Duration(exposeSec*1000)*time.Millisecond + 30*time.Second
	case "DITHERING":
		return 30 * time.Second
	case "PARKING":
		return 90 * time.Second
	default:
		return 30 * time.Second
	}
}
