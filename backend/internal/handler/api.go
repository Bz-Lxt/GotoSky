package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gotosky/gotosky/internal/astro"
	"github.com/gotosky/gotosky/internal/config"
	"github.com/gotosky/gotosky/internal/domain"
	"github.com/gotosky/gotosky/internal/engine"
	"github.com/gotosky/gotosky/internal/httpx"
	"github.com/gotosky/gotosky/internal/service"
	"github.com/gotosky/gotosky/internal/store"
	"github.com/gotosky/gotosky/internal/telescope"
	"github.com/gotosky/gotosky/internal/timeutil"
	"github.com/gotosky/gotosky/internal/weather"
	"github.com/gotosky/gotosky/internal/ws"
	"golang.org/x/crypto/bcrypt"
)

type API struct {
	Cfg     config.Config
	Store   *store.Store
	Scorer  *service.Scorer
	Weather weather.Provider
	Guard   *weather.Guard
	Hub     *ws.Hub
	Reg     *telescope.Registry
	NewDrv  func() telescope.Driver
}

func (a *API) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(httpx.RequestID, httpx.Recoverer, httpx.CORS(a.Cfg.CORSOrigins))
	r.Get("/health", a.health)
	r.Post("/api/v1/auth/login", a.login)
	r.Get("/api/v1/meta", a.meta)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/sites", a.listSites)
		r.Post("/sites", a.createSite)
		r.Get("/sites/{id}", a.getSite)
		r.Put("/sites/{id}", a.updateSite)
		r.Delete("/sites/{id}", a.deleteSite)

		r.Get("/targets", a.listTargets)
		r.Get("/targets/{id}", a.getTarget)

		r.Get("/equipment", a.listEq)
		r.Post("/equipment", a.createEq)
		r.Get("/rigs", a.listRigs)
		r.Post("/rigs", a.createRig)

		r.Get("/scores", a.scores)
		r.Get("/windows", a.windows)
		r.Post("/forecast/refresh", a.refresh)
		r.Get("/skytrack", a.skytrack)
		r.Get("/night", a.night)
		r.Get("/mystery", a.mystery)
		r.Get("/heatmap", a.heatmap)

		r.Get("/plans", a.listPlans)
		r.Post("/plans", a.createPlan)
		r.Post("/plans/{id}/items", a.addItem)

		r.Get("/sessions", a.listSessions)
		r.Post("/sessions", a.createSession)
		r.Post("/sessions/{id}/commands", a.cmdSession)
		r.Get("/alerts", a.alerts)
		r.Get("/quota", a.quota)
		r.Get("/stars", a.stars)
		r.Get("/riseset", a.riseset)
	})
	r.Get("/ws", a.Hub.Handle)
	return r
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	httpx.Raw(w, 200, map[string]any{"status": "ok", "engine": engine.EngineVersion, "tz": "Asia/Shanghai"})
}

func (a *API) meta(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, 200, map[string]any{
		"weather_provider": a.Cfg.WeatherProvider,
		"telescope_driver": a.Cfg.TelescopeDriver,
		"engine_version":   engine.EngineVersion,
		"seeing_is_derived": true,
		"quota_remaining":  remaining(a.Guard),
	})
}

func remaining(g *weather.Guard) int {
	if g == nil {
		return 0
	}
	return g.Remaining()
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var req struct{ Username, Password string }
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, 400, "invalid_json", "invalid json")
		return
	}
	u, err := a.Store.UserByName(r.Context(), req.Username)
	if err != nil {
		httpx.Error(w, 401, "unauthorized", "bad credentials")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		httpx.Error(w, 401, "unauthorized", "bad credentials")
		return
	}
	tok := httpx.SignToken(a.Cfg.JWTSecret, u.Username, 24*time.Hour)
	httpx.JSON(w, 200, map[string]any{"token": tok, "user": u})
}

func (a *API) listSites(w http.ResponseWriter, r *http.Request) {
	xs, err := a.Store.ListSites(r.Context())
	if err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	httpx.JSON(w, 200, xs)
}

func parseSite(r *http.Request) (domain.Site, error) {
	var s domain.Site
	if err := httpx.Decode(r, &s); err != nil {
		return s, err
	}
	if s.Name == "" || s.Latitude < -90 || s.Latitude > 90 || s.Longitude < -180 || s.Longitude > 180 {
		return s, errRange
	}
	if s.Bortle < 1 || s.Bortle > 9 {
		return s, errRange
	}
	if s.SQM == 0 {
		s.SQM = engine.BortleToSQM(s.Bortle)
	}
	if s.Timezone == "" {
		s.Timezone = "Asia/Shanghai"
	}
	if s.MinAltitude == 0 {
		s.MinAltitude = 20
	}
	if s.HorizonMask == nil {
		s.HorizonMask = domain.EmptyJSONArray()
	}
	return s, nil
}

var errRange = jsonError("out of range")

type jsonError string

func (e jsonError) Error() string { return string(e) }

func (a *API) createSite(w http.ResponseWriter, r *http.Request) {
	s, err := parseSite(r)
	if err != nil {
		httpx.Error(w, 422, "validation_error", err.Error())
		return
	}
	s.ID = uuid.New()
	if err := a.Store.UpsertSite(r.Context(), s); err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	w.Header().Set("Location", "/api/v1/sites/"+s.ID.String())
	httpx.JSON(w, 201, s)
}

func (a *API) getSite(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, 400, "bad_id", "bad id")
		return
	}
	s, err := a.Store.GetSite(r.Context(), id)
	if err != nil {
		httpx.Error(w, 404, "not_found", "site not found")
		return
	}
	httpx.JSON(w, 200, s)
}

func (a *API) updateSite(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, 400, "bad_id", "bad id")
		return
	}
	s, err := parseSite(r)
	if err != nil {
		httpx.Error(w, 422, "validation_error", err.Error())
		return
	}
	s.ID = id
	if err := a.Store.UpsertSite(r.Context(), s); err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	httpx.JSON(w, 200, s)
}

func (a *API) deleteSite(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, 400, "bad_id", "bad id")
		return
	}
	_ = a.Store.DeleteSite(r.Context(), id)
	w.WriteHeader(204)
}

func (a *API) listTargets(w http.ResponseWriter, r *http.Request) {
	xs, err := a.Store.ListTargets(r.Context(), r.URL.Query().Get("q"), r.URL.Query().Get("kind"))
	if err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	httpx.JSON(w, 200, xs)
}

func (a *API) getTarget(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	t, err := a.Store.GetTarget(r.Context(), id)
	if err != nil {
		httpx.Error(w, 404, "not_found", "target not found")
		return
	}
	httpx.JSON(w, 200, t)
}

func (a *API) listEq(w http.ResponseWriter, r *http.Request) {
	xs, err := a.Store.ListEquipment(r.Context())
	if err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	httpx.JSON(w, 200, xs)
}

func (a *API) createEq(w http.ResponseWriter, r *http.Request) {
	var e domain.Equipment
	if err := httpx.Decode(r, &e); err != nil {
		httpx.Error(w, 400, "invalid_json", "invalid json")
		return
	}
	e.Kind = strings.ToUpper(e.Kind)
	if e.Name == "" {
		httpx.Error(w, 422, "validation_error", "name required")
		return
	}
	e.ID = uuid.New()
	if err := a.Store.UpsertEquipment(r.Context(), e); err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	httpx.JSON(w, 201, e)
}

func (a *API) listRigs(w http.ResponseWriter, r *http.Request) {
	xs, err := a.Store.ListRigs(r.Context())
	if err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	eq, _ := a.Store.ListEquipment(r.Context())
	xs = service.AnnotateFOV(xs, eq)
	httpx.JSON(w, 200, xs)
}

func (a *API) createRig(w http.ResponseWriter, r *http.Request) {
	var rg domain.Rig
	if err := httpx.Decode(r, &rg); err != nil {
		httpx.Error(w, 400, "invalid_json", "invalid json")
		return
	}
	rg.ID = uuid.New()
	if err := a.Store.UpsertRig(r.Context(), rg); err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	httpx.JSON(w, 201, rg)
}

func (a *API) scores(w http.ResponseWriter, r *http.Request) {
	siteID, err := uuid.Parse(r.URL.Query().Get("site_id"))
	if err != nil {
		httpx.Error(w, 400, "validation_error", "site_id required")
		return
	}
	var tgt uuid.UUID
	if q := r.URL.Query().Get("target_id"); q != "" {
		tgt, _ = uuid.Parse(q)
	}
	from := time.Now().UTC().Add(-1 * time.Hour)
	to := from.Add(8 * 24 * time.Hour)
	xs, err := a.Store.ListScores(r.Context(), siteID, tgt, from, to)
	if err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	httpx.JSON(w, 200, xs)
}

func (a *API) windows(w http.ResponseWriter, r *http.Request) {
	siteID, err := uuid.Parse(r.URL.Query().Get("site_id"))
	if err != nil {
		httpx.Error(w, 400, "validation_error", "site_id required")
		return
	}
	xs, err := a.Store.ListWindows(r.Context(), siteID)
	if err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	httpx.JSON(w, 200, xs)
}

func (a *API) refresh(w http.ResponseWriter, r *http.Request) {
	var req struct{ SiteID uuid.UUID `json:"site_id"` }
	if err := httpx.Decode(r, &req); err != nil || req.SiteID == uuid.Nil {
		httpx.Error(w, 422, "validation_error", "site_id required")
		return
	}
	site, err := a.Store.GetSite(r.Context(), req.SiteID)
	if err != nil {
		httpx.Error(w, 404, "not_found", "site")
		return
	}
	tgts, err := a.Store.ListTargets(r.Context(), "", "")
	if err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	if len(tgts) > 40 {
		tgts = tgts[:40]
	}
	if err := a.Scorer.Recompute(r.Context(), site, tgts, 7); err != nil {
		httpx.Error(w, 502, "weather", err.Error())
		return
	}
	httpx.JSON(w, 200, map[string]any{"ok": true, "quota_remaining": remaining(a.Guard)})
}

func (a *API) skytrack(w http.ResponseWriter, r *http.Request) {
	siteID, _ := uuid.Parse(r.URL.Query().Get("site_id"))
	tgtID, _ := uuid.Parse(r.URL.Query().Get("target_id"))
	site, err := a.Store.GetSite(r.Context(), siteID)
	if err != nil {
		httpx.Error(w, 404, "not_found", "site")
		return
	}
	tgt, err := a.Store.GetTarget(r.Context(), tgtID)
	if err != nil {
		httpx.Error(w, 404, "not_found", "target")
		return
	}
	from := time.Now().UTC().Truncate(time.Hour)
	httpx.JSON(w, 200, a.Scorer.SkyTrack(site, tgt, from, 24))
}

func (a *API) night(w http.ResponseWriter, r *http.Request) {
	siteID, _ := uuid.Parse(r.URL.Query().Get("site_id"))
	site, err := a.Store.GetSite(r.Context(), siteID)
	if err != nil {
		httpx.Error(w, 404, "not_found", "site")
		return
	}
	loc := timeutil.ParseLocation(site.Timezone)
	ev := astro.NightEvents(time.Now(), site.Latitude, site.Longitude, loc)
	httpx.JSON(w, 200, ev)
}

func (a *API) mystery(w http.ResponseWriter, r *http.Request) {
	siteID, _ := uuid.Parse(r.URL.Query().Get("site_id"))
	wins, err := a.Store.ListWindows(r.Context(), siteID)
	if err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	tgts, _ := a.Store.ListTargets(r.Context(), "", "")
	byID := map[uuid.UUID]domain.Target{}
	for _, t := range tgts {
		byID[t.ID] = t
	}
	cands := []engine.Candidate{}
	for _, w := range wins {
		t := byID[w.TargetID]
		sz := 20.0
		if t.SizeArcmin != nil {
			sz = *t.SizeArcmin
		}
		cands = append(cands, engine.Candidate{
			Target: t, MeanScore: w.MeanScore, PeakScore: w.PeakScore, FOVFit: engine.FOVFit(sz, 72),
		})
	}
	httpx.JSON(w, 200, engine.Recommend(cands, 3))
}

func (a *API) heatmap(w http.ResponseWriter, r *http.Request) {
	a.scores(w, r)
}

func (a *API) listPlans(w http.ResponseWriter, r *http.Request) {
	var site uuid.UUID
	if q := r.URL.Query().Get("site_id"); q != "" {
		site, _ = uuid.Parse(q)
	}
	xs, err := a.Store.ListPlans(r.Context(), site)
	if err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	httpx.JSON(w, 200, xs)
}

func (a *API) createPlan(w http.ResponseWriter, r *http.Request) {
	var p domain.Plan
	if err := httpx.Decode(r, &p); err != nil || p.SiteID == uuid.Nil || p.Title == "" {
		httpx.Error(w, 422, "validation_error", "site_id and title required")
		return
	}
	p.ID = uuid.New()
	if err := a.Store.UpsertPlan(r.Context(), p); err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	httpx.JSON(w, 201, p)
}

func (a *API) addItem(w http.ResponseWriter, r *http.Request) {
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, 400, "bad_id", "bad id")
		return
	}
	var it domain.PlanItem
	if err := httpx.Decode(r, &it); err != nil {
		httpx.Error(w, 400, "invalid_json", "invalid json")
		return
	}
	it.ID = uuid.New()
	it.PlanID = pid
	if it.ExposureS <= 0 {
		it.ExposureS = 180
	}
	if it.FrameCount <= 0 {
		it.FrameCount = 10
	}
	if it.StartUTC.IsZero() || it.EndUTC.IsZero() || it.EndUTC.Before(it.StartUTC) {
		httpx.Error(w, 422, "validation_error", "invalid time range")
		return
	}
	if err := a.Store.AddPlanItem(r.Context(), it); err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	httpx.JSON(w, 201, it)
}

func (a *API) listSessions(w http.ResponseWriter, r *http.Request) {
	live := a.Reg.Snapshots()
	if len(live) > 0 {
		httpx.JSON(w, 200, live)
		return
	}
	xs, err := a.Store.ListSessions(r.Context())
	if err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	httpx.JSON(w, 200, xs)
}

func (a *API) createSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RigID      uuid.UUID  `json:"rig_id"`
		PlanItemID *uuid.UUID `json:"plan_item_id"`
		RA         float64    `json:"ra"`
		Dec        float64    `json:"dec"`
		Frames     int        `json:"frames"`
		ExposureS  float64    `json:"exposure_s"`
	}
	if err := httpx.Decode(r, &req); err != nil || req.RigID == uuid.Nil {
		httpx.Error(w, 422, "validation_error", "rig_id required")
		return
	}
	sess := domain.Session{
		ID: uuid.New(), RigID: req.RigID, PlanItemID: req.PlanItemID,
		State: "IDLE", SourceMode: sourceMode(a.Cfg), ProgressN: req.Frames,
	}
	if err := a.Store.SaveSession(r.Context(), sess); err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	act := a.Reg.Attach(r.Context(), sess)
	cmdID := uuid.New()
	body := map[string]any{"ra": req.RA, "dec": req.Dec, "frames": float64(max(req.Frames, 4)), "exposure_s": req.ExposureS}
	go func() { _ = act.Submit(cmdID, "START", body) }()
	httpx.JSON(w, 201, sess)
}

func (a *API) cmdSession(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, 400, "bad_id", "bad id")
		return
	}
	var req struct {
		CommandID uuid.UUID      `json:"command_id"`
		Verb      string         `json:"verb"`
		Payload   map[string]any `json:"payload"`
	}
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, 400, "invalid_json", "invalid json")
		return
	}
	if req.CommandID == uuid.Nil {
		req.CommandID = uuid.New()
	}
	req.Verb = strings.ToUpper(req.Verb)
	act := a.Reg.Get(id)
	if act == nil {
		httpx.Error(w, 404, "not_found", "session not live")
		return
	}
	if err := act.Submit(req.CommandID, req.Verb, req.Payload); err != nil {
		httpx.Error(w, 409, "conflict", err.Error())
		return
	}
	httpx.JSON(w, 200, map[string]any{"command_id": req.CommandID, "ok": true})
}

func (a *API) alerts(w http.ResponseWriter, r *http.Request) {
	xs, err := a.Store.ListAlerts(r.Context())
	if err != nil {
		httpx.Error(w, 500, "db", err.Error())
		return
	}
	httpx.JSON(w, 200, xs)
}

func (a *API) stars(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, 200, astro.SkyField())
}

func (a *API) riseset(w http.ResponseWriter, r *http.Request) {
	siteID, _ := uuid.Parse(r.URL.Query().Get("site_id"))
	tgtID, _ := uuid.Parse(r.URL.Query().Get("target_id"))
	site, err := a.Store.GetSite(r.Context(), siteID)
	if err != nil {
		httpx.Error(w, 404, "not_found", "site")
		return
	}
	tgt, err := a.Store.GetTarget(r.Context(), tgtID)
	if err != nil {
		httpx.Error(w, 404, "not_found", "target")
		return
	}
	loc := timeutil.ParseLocation(site.Timezone)
	httpx.JSON(w, 200, astro.ObjectRiseSet(time.Now(), site.Latitude, site.Longitude, tgt.RAHours, tgt.DecDeg, site.MinAltitude, loc))
}

func (a *API) quota(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, 200, map[string]any{
		"daily_limit": a.Cfg.WeatherQuota, "remaining": remaining(a.Guard),
		"unit_cost_cny": "0", "estimated_next_call": "¥0",
	})
}

func sourceMode(c config.Config) string {
	if c.TelescopeIsMock() {
		return "SIMULATED"
	}
	return "DEVICE"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
