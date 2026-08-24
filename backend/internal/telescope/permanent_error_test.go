package telescope_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/gotosky/gotosky/internal/domain"
	"github.com/gotosky/gotosky/internal/telescope"
)

type permanentConnectDriver struct {
	mu    sync.Mutex
	calls []string
}

func (d *permanentConnectDriver) record(call string) {
	d.mu.Lock()
	d.calls = append(d.calls, call)
	d.mu.Unlock()
}

func (d *permanentConnectDriver) Calls() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]string(nil), d.calls...)
}

func (d *permanentConnectDriver) Name() string   { return "alpaca" }
func (d *permanentConnectDriver) Source() string { return "DEVICE" }
func (d *permanentConnectDriver) Inject(string)  {}

func (d *permanentConnectDriver) Connect(context.Context) error {
	d.record("connect")
	return &telescope.DeviceError{
		Code:  "AuthFailed",
		Class: telescope.ClassPermanent,
		Msg:   "credentials rejected",
	}
}

func (d *permanentConnectDriver) Disconnect(context.Context) error { return nil }
func (d *permanentConnectDriver) Slew(context.Context, float64, float64) error {
	d.record("slew")
	return nil
}
func (d *permanentConnectDriver) WaitSlew(context.Context) error {
	d.record("wait_slew")
	return nil
}
func (d *permanentConnectDriver) Settle(context.Context) error {
	d.record("settle")
	return nil
}
func (d *permanentConnectDriver) LockGuide(context.Context) error {
	d.record("guide")
	return nil
}
func (d *permanentConnectDriver) SetFilter(context.Context, int) error {
	d.record("filter")
	return nil
}
func (d *permanentConnectDriver) Expose(context.Context, float64) error {
	d.record("expose")
	return nil
}
func (d *permanentConnectDriver) Dither(context.Context) error {
	d.record("dither")
	return nil
}
func (d *permanentConnectDriver) Park(context.Context) error {
	d.record("park")
	return nil
}
func (d *permanentConnectDriver) Heartbeat(context.Context) error { return nil }
func (d *permanentConnectDriver) ReadSensors(context.Context) (telescope.Sensors, error) {
	return telescope.Sensors{Source: d.Source()}, nil
}

func TestSessionStopsAfterPermanentConnectFailure(t *testing.T) {
	drv := &permanentConnectDriver{}
	sess := domain.Session{
		ID:         uuid.New(),
		RigID:      uuid.New(),
		State:      "IDLE",
		SourceMode: "DEVICE",
	}
	actor := telescope.NewActor(sess, drv, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	actor.Start(ctx)
	defer actor.Stop()

	_ = actor.Submit(uuid.New(), "START", map[string]any{
		"ra":         1.25,
		"dec":        22.5,
		"frames":     1.0,
		"exposure_s": 0.1,
	})

	got := actor.Snapshot()
	if got.State != "ERROR" {
		t.Fatalf("session state = %q, want ERROR", got.State)
	}
	if !strings.Contains(got.LastError, "AuthFailed") {
		t.Fatalf("last error = %q, want AuthFailed", got.LastError)
	}
	if calls := drv.Calls(); strings.Join(calls, ",") != "connect" {
		t.Fatalf("driver calls after permanent connect failure = %v, want [connect]", calls)
	}
}
