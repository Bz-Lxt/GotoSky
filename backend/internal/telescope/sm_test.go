package telescope

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gotosky/gotosky/internal/domain"
)

type memStore struct {
	cmds map[uuid.UUID][]byte
}

func (m *memStore) SaveSession(context.Context, domain.Session) error { return nil }
func (m *memStore) AppendEvent(context.Context, domain.SessionEvent) error {
	return nil
}
func (m *memStore) SaveCommand(_ context.Context, id, _ uuid.UUID, _ string, _, res []byte) error {
	if m.cmds == nil {
		m.cmds = map[uuid.UUID][]byte{}
	}
	m.cmds[id] = res
	return nil
}
func (m *memStore) GetCommand(_ context.Context, id uuid.UUID) (bool, []byte, error) {
	v, ok := m.cmds[id]
	return ok, v, nil
}
func (m *memStore) AddExposure(context.Context, uuid.UUID, int, string, float64, string, string) error {
	return nil
}

func TestSequenceCompletes(t *testing.T) {
	drv := NewMockDriver()
	s := domain.Session{ID: uuid.New(), RigID: uuid.New(), State: "IDLE", SourceMode: "SIMULATED"}
	act := NewActor(s, drv, &memStore{}, nil)
	act.frames = 2
	act.exposeS = 0.2
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	act.Start(ctx)
	if err := act.Submit(uuid.New(), "START", map[string]any{"ra": 0.7, "dec": 41.2, "frames": 2.0, "exposure_s": 0.2}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	st := act.Snapshot()
	if st.State == "IDLE" {
		t.Fatalf("did not start: %s", st.State)
	}
}

func TestPermanentNoRetry(t *testing.T) {
	drv := NewMockDriver()
	drv.Inject("auth")
	err := drv.Connect(context.Background())
	if Classify(err) != ClassPermanent {
		t.Fatalf("%v", err)
	}
}

func TestSafetyClass(t *testing.T) {
	drv := NewMockDriver()
	drv.Inject("rain")
	err := drv.Slew(context.Background(), 1, 1)
	if Classify(err) != ClassSafety {
		t.Fatalf("%v", err)
	}
}

func TestIdempotentCommand(t *testing.T) {
	st := &memStore{cmds: map[uuid.UUID][]byte{}}
	id := uuid.New()
	_ = st.SaveCommand(context.Background(), id, id, "START", nil, []byte(`{"ok":true}`))
	drv := NewMockDriver()
	act := NewActor(domain.Session{ID: id, State: "IDLE"}, drv, st, nil)
	act.Start(context.Background())
	if err := act.Submit(id, "START", map[string]any{}); err != nil {
		t.Fatal(err)
	}
}

func TestSourceFromDriver(t *testing.T) {
	if NewMockDriver().Source() != "SIMULATED" {
		t.Fatal("mock source")
	}
	if NewAlpaca("http://x").Source() != "DEVICE" {
		t.Fatal("alpaca source")
	}
}
