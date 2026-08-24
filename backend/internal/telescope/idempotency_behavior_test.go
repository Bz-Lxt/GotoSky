package telescope_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/gotosky/gotosky/internal/domain"
	"github.com/gotosky/gotosky/internal/telescope"
)

type commandMemory struct {
	mu      sync.Mutex
	results map[uuid.UUID][]byte
}

func (m *commandMemory) SaveSession(context.Context, domain.Session) error { return nil }

func (m *commandMemory) AppendEvent(context.Context, domain.SessionEvent) error { return nil }

func (m *commandMemory) SaveCommand(_ context.Context, id, _ uuid.UUID, _ string, _, result []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.results[id] = append([]byte(nil), result...)
	return nil
}

func (m *commandMemory) GetCommand(_ context.Context, id uuid.UUID) (bool, []byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result, ok := m.results[id]
	return ok, append([]byte(nil), result...), nil
}

func (m *commandMemory) AddExposure(context.Context, uuid.UUID, int, string, float64, string, string) error {
	return nil
}

type recordingDriver struct {
	mu    sync.Mutex
	calls map[string]int
}

func (d *recordingDriver) record(operation string) error {
	d.mu.Lock()
	d.calls[operation]++
	d.mu.Unlock()
	return nil
}

func (d *recordingDriver) callCount(operation string) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.calls[operation]
}

func (d *recordingDriver) Name() string   { return "recording" }
func (d *recordingDriver) Source() string { return "SIMULATED" }
func (d *recordingDriver) Connect(context.Context) error {
	return d.record("connect")
}
func (d *recordingDriver) Disconnect(context.Context) error {
	return d.record("disconnect")
}
func (d *recordingDriver) Slew(context.Context, float64, float64) error {
	return d.record("slew")
}
func (d *recordingDriver) WaitSlew(context.Context) error {
	return d.record("wait_slew")
}
func (d *recordingDriver) Settle(context.Context) error {
	return d.record("settle")
}
func (d *recordingDriver) LockGuide(context.Context) error {
	return d.record("lock_guide")
}
func (d *recordingDriver) SetFilter(context.Context, int) error {
	return d.record("set_filter")
}
func (d *recordingDriver) Expose(context.Context, float64) error {
	return d.record("expose")
}
func (d *recordingDriver) Dither(context.Context) error {
	return d.record("dither")
}
func (d *recordingDriver) Park(context.Context) error {
	return d.record("park")
}
func (d *recordingDriver) Heartbeat(context.Context) error {
	return d.record("heartbeat")
}
func (d *recordingDriver) ReadSensors(context.Context) (telescope.Sensors, error) {
	return telescope.Sensors{Source: d.Source()}, nil
}
func (d *recordingDriver) Inject(string) {}

func TestDuplicateStartCommandDoesNotRepeatDeviceSequence(t *testing.T) {
	driver := &recordingDriver{calls: make(map[string]int)}
	persist := &commandMemory{results: make(map[uuid.UUID][]byte)}
	session := domain.Session{ID: uuid.New(), RigID: uuid.New(), State: "IDLE", SourceMode: "SIMULATED"}
	actor := telescope.NewActor(session, driver, persist, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	actor.Start(ctx)

	commandID := uuid.New()
	payload := map[string]any{"ra": 1.25, "dec": 30.0, "frames": 1.0, "exposure_s": 0.1}
	for attempt := 1; attempt <= 2; attempt++ {
		if err := actor.Submit(commandID, "START", payload); err != nil {
			t.Fatalf("attempt %d: %v", attempt, err)
		}
	}

	for _, operation := range []string{"connect", "slew", "lock_guide", "expose", "park"} {
		if got := driver.callCount(operation); got != 1 {
			t.Errorf("%s called %d times, want 1 for a replayed command", operation, got)
		}
	}
}
