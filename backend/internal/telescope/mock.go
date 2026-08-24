package telescope

import (
	"context"
	"math"
	"sync"
	"time"
)

type MockDriver struct {
	mu       sync.Mutex
	fault    string
	ra, dec  float64
	filter   int
	src      string
	tick     time.Duration
}

func NewMockDriver() *MockDriver {
	return &MockDriver{src: "SIMULATED", tick: 40 * time.Millisecond}
}

func (m *MockDriver) Name() string   { return "mock" }
func (m *MockDriver) Source() string { return m.src }
func (m *MockDriver) Inject(fault string) {
	m.mu.Lock()
	m.fault = fault
	m.mu.Unlock()
}

func (m *MockDriver) popFault(op string) error {
	m.mu.Lock()
	f := m.fault
	if f != "" {
		m.fault = ""
	}
	m.mu.Unlock()
	switch f {
	case "timeout":
		return &DeviceError{Code: "DeviceTimeout", Class: ClassTransient, Msg: op + " timeout"}
	case "busy":
		return &DeviceError{Code: "DeviceBusy", Class: ClassTransient, Msg: op + " busy"}
	case "reset":
		return &DeviceError{Code: "ConnectionReset", Class: ClassTransient, Msg: "link reset"}
	case "auth":
		return &DeviceError{Code: "AuthFailed", Class: ClassPermanent, Msg: "auth failed"}
	case "invalid":
		return &DeviceError{Code: "InvalidCommand", Class: ClassPermanent, Msg: "bad command"}
	case "horizon":
		return &DeviceError{Code: "TargetBelowHorizon", Class: ClassPermanent, Msg: "below horizon"}
	case "hardware":
		return &DeviceError{Code: "HardwareFault", Class: ClassPermanent, Msg: "hardware"}
	case "rain":
		return &DeviceError{Code: "RainDetected", Class: ClassSafety, Msg: "rain"}
	case "wind":
		return &DeviceError{Code: "WindLimitExceeded", Class: ClassSafety, Msg: "wind"}
	case "estop":
		return &DeviceError{Code: "EmergencyStop", Class: ClassSafety, Msg: "estop"}
	}
	return nil
}

func (m *MockDriver) sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (m *MockDriver) Connect(ctx context.Context) error {
	if err := m.popFault("connect"); err != nil {
		return err
	}
	return m.sleep(ctx, 80*time.Millisecond)
}
func (m *MockDriver) Disconnect(_ context.Context) error { return nil }

func (m *MockDriver) Slew(ctx context.Context, ra, dec float64) error {
	if err := m.popFault("slew"); err != nil {
		return err
	}
	m.mu.Lock()
	m.ra, m.dec = ra, dec
	m.mu.Unlock()
	return m.sleep(ctx, 200*time.Millisecond)
}
func (m *MockDriver) WaitSlew(ctx context.Context) error { return m.sleep(ctx, 80*time.Millisecond) }
func (m *MockDriver) Settle(ctx context.Context) error {
	if err := m.popFault("settle"); err != nil {
		return err
	}
	return m.sleep(ctx, 80*time.Millisecond)
}
func (m *MockDriver) LockGuide(ctx context.Context) error {
	if err := m.popFault("guide"); err != nil {
		return err
	}
	return m.sleep(ctx, 100*time.Millisecond)
}
func (m *MockDriver) SetFilter(ctx context.Context, pos int) error {
	if err := m.popFault("filter"); err != nil {
		return err
	}
	m.mu.Lock()
	m.filter = pos
	m.mu.Unlock()
	return m.sleep(ctx, 60*time.Millisecond)
}
func (m *MockDriver) Expose(ctx context.Context, seconds float64) error {
	if err := m.popFault("expose"); err != nil {
		return err
	}
	d := time.Duration(seconds * float64(time.Second))
	if d > 2*time.Second {
		d = 2 * time.Second // demo speed
	}
	if d < 200*time.Millisecond {
		d = 200 * time.Millisecond
	}
	return m.sleep(ctx, d)
}
func (m *MockDriver) Dither(ctx context.Context) error {
	if err := m.popFault("dither"); err != nil {
		return err
	}
	return m.sleep(ctx, 80*time.Millisecond)
}
func (m *MockDriver) Park(ctx context.Context) error {
	if err := m.popFault("park"); err != nil {
		return err
	}
	return m.sleep(ctx, 120*time.Millisecond)
}
func (m *MockDriver) Heartbeat(ctx context.Context) error {
	if err := m.popFault("hb"); err != nil {
		return err
	}
	return m.sleep(ctx, 10*time.Millisecond)
}
func (m *MockDriver) ReadSensors(_ context.Context) (Sensors, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := float64(time.Now().UnixNano()%100000) / 100000
	return Sensors{
		MountRA: m.ra, MountDec: m.dec, Slewing: false,
		GuideRMS: 0.35 + 0.05*math.Sin(now*6), CCDTemp: -10 + now,
		FilterPos: m.filter, Humidity: 42 + now*3, WindMS: 2.1, Source: m.src,
	}, nil
}
