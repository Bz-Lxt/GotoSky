package telescope

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/gotosky/gotosky/internal/domain"
)

// Registry holds live actors. NEVER lock an actor while holding r.mu.
type Registry struct {
	mu      sync.Mutex
	actors  map[uuid.UUID]*SessionActor
	factory func(domain.Session) *SessionActor
}

func NewRegistry(factory func(domain.Session) *SessionActor) *Registry {
	return &Registry{actors: map[uuid.UUID]*SessionActor{}, factory: factory}
}

func (r *Registry) Attach(ctx context.Context, s domain.Session) *SessionActor {
	act := r.factory(s)
	r.mu.Lock()
	r.actors[s.ID] = act
	r.mu.Unlock()
	act.Start(ctx)
	return act
}

func (r *Registry) Get(id uuid.UUID) *SessionActor {
	r.mu.Lock()
	a := r.actors[id]
	r.mu.Unlock()
	return a
}

func (r *Registry) Live() []*SessionActor {
	r.mu.Lock()
	out := make([]*SessionActor, 0, len(r.actors))
	for _, a := range r.actors {
		out = append(out, a)
	}
	r.mu.Unlock()
	return out
}

func (r *Registry) Snapshots() []domain.Session {
	live := r.Live()
	out := make([]domain.Session, 0, len(live))
	for _, a := range live {
		out = append(out, a.Snapshot())
	}
	return out
}
