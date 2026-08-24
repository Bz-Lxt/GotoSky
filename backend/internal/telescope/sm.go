package telescope

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gotosky/gotosky/internal/domain"
)

type Persist interface {
	SaveSession(ctx context.Context, s domain.Session) error
	AppendEvent(ctx context.Context, ev domain.SessionEvent) error
	SaveCommand(ctx context.Context, commandID, sessionID uuid.UUID, verb string, payload, result []byte) error
	GetCommand(ctx context.Context, commandID uuid.UUID) (found bool, result []byte, err error)
	AddExposure(ctx context.Context, sessionID uuid.UUID, seq int, filter string, dur float64, status, filename string) error
}

type Broadcaster interface {
	Publish(msg domain.Telemetry)
}

type SessionActor struct {
	mu       sync.Mutex
	sess     domain.Session
	drv      Driver
	store    Persist
	bus      Broadcaster
	cmds     chan command
	cancel   context.CancelFunc
	hbMiss   int
	ra, dec  float64
	frames   int
	exposeS  float64
	filters  []int
}

type command struct {
	id   uuid.UUID
	verb string
	body map[string]any
	res  chan error
}

func NewActor(s domain.Session, drv Driver, store Persist, bus Broadcaster) *SessionActor {
	return &SessionActor{
		sess:    s,
		drv:     drv,
		store:   store,
		bus:     bus,
		cmds:    make(chan command, 16),
		exposeS: 30,
		frames:  6,
		filters: []int{0},
	}
}

func (a *SessionActor) Start(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	a.cancel = cancel
	go a.loop(ctx)
}

func (a *SessionActor) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
}

func (a *SessionActor) Snapshot() domain.Session {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.sess
}

func (a *SessionActor) Submit(id uuid.UUID, verb string, body map[string]any) error {
	res := make(chan error, 1)
	select {
	case a.cmds <- command{id: id, verb: verb, body: body, res: res}:
	default:
		return &DeviceError{Code: "DeviceBusy", Class: ClassTransient, Msg: "queue full"}
	}
	return <-res
}

func (a *SessionActor) loop(ctx context.Context) {
	hb := time.NewTicker(5 * time.Second)
	defer hb.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-a.cmds:
			err := a.handle(ctx, cmd)
			cmd.res <- err
		case <-hb.C:
			a.watchdog(ctx)
		}
	}
}

func (a *SessionActor) handle(ctx context.Context, cmd command) error {
	if a.store != nil && cmd.id != uuid.Nil {
		found, _, err := a.store.GetCommand(ctx, cmd.id)
		if err == nil && found {
			return nil
		}
	}
	var runErr error
	switch cmd.verb {
	case "START":
		if v, ok := cmd.body["ra"].(float64); ok {
			a.ra = v
		}
		if v, ok := cmd.body["dec"].(float64); ok {
			a.dec = v
		}
		if v, ok := cmd.body["frames"].(float64); ok {
			a.frames = int(v)
		}
		if v, ok := cmd.body["exposure_s"].(float64); ok && v > 0 {
			a.exposeS = v
		}
		runErr = a.runSequence(ctx)
	case "ABORT":
		runErr = a.transition(ctx, "ABORTED", "PERMANENT", "abort")
	case "INJECT":
		if f, _ := cmd.body["fault"].(string); f != "" {
			a.drv.Inject(f)
		}
	default:
		runErr = &DeviceError{Code: "InvalidCommand", Class: ClassPermanent, Msg: cmd.verb}
	}
	if a.store != nil && cmd.id != uuid.Nil {
		raw, _ := json.Marshal(cmd.body)
		res, _ := json.Marshal(map[string]any{"ok": runErr == nil})
		_ = a.store.SaveCommand(ctx, cmd.id, a.sess.ID, cmd.verb, raw, res)
	}
	return runErr
}

func (a *SessionActor) runSequence(ctx context.Context) error {
	steps := []struct {
		state string
		fn    func(context.Context) error
	}{
		{"CONNECTING", a.drv.Connect},
		{"CONNECTED", func(ctx context.Context) error { return a.transition(ctx, "CONNECTED", "", "connected") }},
		{"SLEWING", func(ctx context.Context) error {
			if err := a.drv.Slew(ctx, a.ra, a.dec); err != nil {
				return err
			}
			return a.drv.WaitSlew(ctx)
		}},
		{"SETTLING", a.drv.Settle},
		{"GUIDE_LOCKING", a.drv.LockGuide},
		{"GUIDING", func(ctx context.Context) error { return a.transition(ctx, "GUIDING", "", "locked") }},
	}
	for _, st := range steps {
		if err := a.execState(ctx, st.state, st.fn); err != nil {
			return a.onFail(ctx, err)
		}
	}
	a.mu.Lock()
	a.sess.ProgressN = a.frames
	a.mu.Unlock()
	for i := 0; i < a.frames; i++ {
		pos := a.filters[i%len(a.filters)]
		if err := a.execState(ctx, "FILTER_CHANGING", func(ctx context.Context) error { return a.drv.SetFilter(ctx, pos) }); err != nil {
			return a.onFail(ctx, err)
		}
		if err := a.execState(ctx, "EXPOSING", func(ctx context.Context) error { return a.drv.Expose(ctx, a.exposeS) }); err != nil {
			return a.onFail(ctx, err)
		}
		a.mu.Lock()
		a.sess.ProgressK = i + 1
		a.mu.Unlock()
		if a.store != nil {
			_ = a.store.AddExposure(ctx, a.sess.ID, i+1, "L", a.exposeS, "DONE", "sim_"+a.sess.ID.String()+"_"+itoa(i+1)+".fits")
		}
		if i < a.frames-1 {
			if err := a.execState(ctx, "DITHERING", a.drv.Dither); err != nil {
				return a.onFail(ctx, err)
			}
		}
		a.emit()
	}
	if err := a.execState(ctx, "PARKING", a.drv.Park); err != nil {
		return a.onFail(ctx, err)
	}
	return a.transition(ctx, "COMPLETED", "", "done")
}

func (a *SessionActor) execState(ctx context.Context, state string, fn func(context.Context) error) error {
	if err := a.transition(ctx, state, "", ""); err != nil {
		return err
	}
	to := TimeoutOf(state, a.exposeS)
	cctx, cancel := context.WithTimeout(ctx, to)
	defer cancel()
	return retry(cctx, fn)
}

func retry(ctx context.Context, fn func(context.Context) error) error {
	backoff := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second}
	var last error
	for i := 0; i < 4; i++ {
		err := fn(ctx)
		if err == nil {
			return nil
		}
		last = err
		cl := Classify(err)
		if cl == ClassPermanent || cl == ClassSafety {
			return err
		}
		if i == 3 {
			break
		}
		t := time.NewTimer(backoff[min(i, len(backoff)-1)])
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
	}
	return last
}

func (a *SessionActor) onFail(ctx context.Context, err error) error {
	cl := Classify(err)
	if cl == ClassSafety {
		_ = a.transition(ctx, "PARKING", string(cl), err.Error())
		_ = a.drv.Park(ctx)
		return a.transition(ctx, "ERROR", string(cl), err.Error())
	}
	return a.transition(ctx, "ERROR", string(cl), err.Error())
}

func (a *SessionActor) transition(ctx context.Context, to, class, msg string) error {
	a.mu.Lock()
	from := a.sess.State
	a.sess.State = to
	a.sess.LastError = msg
	a.sess.UpdatedAt = time.Now()
	if to == "CONNECTING" && a.sess.StartedAt == nil {
		now := time.Now()
		a.sess.StartedAt = &now
	}
	if to == "COMPLETED" || to == "ERROR" || to == "ABORTED" {
		now := time.Now()
		a.sess.EndedAt = &now
	}
	snap := a.sess
	a.mu.Unlock()
	if a.store != nil {
		_ = a.store.SaveSession(ctx, snap)
		ev := domain.SessionEvent{
			ID: uuid.New(), SessionID: snap.ID, FromState: from, ToState: to, Class: class,
			Context: mustJSON(map[string]any{"msg": msg}), CreatedAt: time.Now(),
		}
		_ = a.store.AppendEvent(ctx, ev)
	}
	a.emit()
	return nil
}

func (a *SessionActor) watchdog(ctx context.Context) {
	err := a.drv.Heartbeat(ctx)
	if err != nil {
		a.hbMiss++
		if a.hbMiss >= 6 { // 30s
			_ = a.transition(ctx, "ERROR", string(ClassTransient), "heartbeat lost")
		}
		return
	}
	a.hbMiss = 0
	a.emit()
}

func (a *SessionActor) emit() {
	if a.bus == nil {
		return
	}
	a.mu.Lock()
	s := a.sess
	a.mu.Unlock()
	sens, _ := a.drv.ReadSensors(context.Background())
	src := a.drv.Source()
	a.bus.Publish(domain.Telemetry{
		SessionID: s.ID, RigID: s.RigID, State: s.State,
		ProgressK: s.ProgressK, ProgressN: s.ProgressN, RemainSec: s.RemainSec,
		Sensors: map[string]any{
			"mount_ra": sens.MountRA, "mount_dec": sens.MountDec, "guide_rms": sens.GuideRMS,
			"ccd_temp": sens.CCDTemp, "filter_pos": sens.FilterPos, "humidity": sens.Humidity, "wind_ms": sens.WindMS,
		},
		Source: src, Alert: s.LastError,
	})
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
