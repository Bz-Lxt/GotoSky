package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gotosky/gotosky/internal/domain"
)

type Store struct{ DB *sql.DB }

func (s *Store) CreateUser(ctx context.Context, u domain.User) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO users (id, username, password_hash, display_name) VALUES ($1,$2,$3,$4)`,
		u.ID, u.Username, u.PasswordHash, u.DisplayName)
	return err
}

func (s *Store) UserByName(ctx context.Context, name string) (domain.User, error) {
	var u domain.User
	err := s.DB.QueryRowContext(ctx, `SELECT id, username, password_hash, display_name FROM users WHERE username=$1`, name).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName)
	return u, err
}

func (s *Store) UpsertSite(ctx context.Context, site domain.Site) error {
	if site.HorizonMask == nil {
		site.HorizonMask = domain.EmptyJSONArray()
	}
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO sites (id,name,latitude,longitude,elevation_m,timezone,bortle,sqm,min_altitude,horizon_mask,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW(),NOW())
		ON CONFLICT (id) DO UPDATE SET name=$2,latitude=$3,longitude=$4,elevation_m=$5,timezone=$6,bortle=$7,sqm=$8,min_altitude=$9,horizon_mask=$10,updated_at=NOW()`,
		site.ID, site.Name, site.Latitude, site.Longitude, site.ElevationM, site.Timezone, site.Bortle, site.SQM, site.MinAltitude, site.HorizonMask)
	return err
}

func (s *Store) ListSites(ctx context.Context) ([]domain.Site, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,name,latitude,longitude,elevation_m,timezone,bortle,sqm,min_altitude,horizon_mask,created_at,updated_at FROM sites ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Site
	for rows.Next() {
		var x domain.Site
		if err := rows.Scan(&x.ID, &x.Name, &x.Latitude, &x.Longitude, &x.ElevationM, &x.Timezone, &x.Bortle, &x.SQM, &x.MinAltitude, &x.HorizonMask, &x.CreatedAt, &x.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	if out == nil {
		out = []domain.Site{}
	}
	return out, rows.Err()
}

func (s *Store) GetSite(ctx context.Context, id uuid.UUID) (domain.Site, error) {
	var x domain.Site
	err := s.DB.QueryRowContext(ctx, `SELECT id,name,latitude,longitude,elevation_m,timezone,bortle,sqm,min_altitude,horizon_mask,created_at,updated_at FROM sites WHERE id=$1`, id).
		Scan(&x.ID, &x.Name, &x.Latitude, &x.Longitude, &x.ElevationM, &x.Timezone, &x.Bortle, &x.SQM, &x.MinAltitude, &x.HorizonMask, &x.CreatedAt, &x.UpdatedAt)
	return x, err
}

func (s *Store) DeleteSite(ctx context.Context, id uuid.UUID) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM sites WHERE id=$1`, id)
	return err
}

func (s *Store) UpsertTarget(ctx context.Context, t domain.Target) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO targets (id,catalog_id,name,name_zh,ra_hours,dec_deg,mag,kind,size_arcmin,notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (catalog_id) DO UPDATE SET name=$3,name_zh=$4,ra_hours=$5,dec_deg=$6,mag=$7,kind=$8,size_arcmin=$9,notes=$10`,
		t.ID, t.CatalogID, t.Name, t.NameZH, t.RAHours, t.DecDeg, t.Mag, t.Kind, t.SizeArcmin, t.Notes)
	return err
}

func (s *Store) ListTargets(ctx context.Context, q, kind string) ([]domain.Target, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id,catalog_id,name,name_zh,ra_hours,dec_deg,mag,kind,size_arcmin,notes FROM targets
		WHERE ($1='' OR catalog_id ILIKE '%'||$1||'%' OR name ILIKE '%'||$1||'%' OR name_zh ILIKE '%'||$1||'%')
		  AND ($2='' OR kind=$2)
		ORDER BY catalog_id`, q, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Target
	for rows.Next() {
		var t domain.Target
		if err := rows.Scan(&t.ID, &t.CatalogID, &t.Name, &t.NameZH, &t.RAHours, &t.DecDeg, &t.Mag, &t.Kind, &t.SizeArcmin, &t.Notes); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if out == nil {
		out = []domain.Target{}
	}
	return out, rows.Err()
}

func (s *Store) GetTarget(ctx context.Context, id uuid.UUID) (domain.Target, error) {
	var t domain.Target
	err := s.DB.QueryRowContext(ctx, `SELECT id,catalog_id,name,name_zh,ra_hours,dec_deg,mag,kind,size_arcmin,notes FROM targets WHERE id=$1`, id).
		Scan(&t.ID, &t.CatalogID, &t.Name, &t.NameZH, &t.RAHours, &t.DecDeg, &t.Mag, &t.Kind, &t.SizeArcmin, &t.Notes)
	return t, err
}

func (s *Store) CountTargets(ctx context.Context) (int, error) {
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM targets`).Scan(&n)
	return n, err
}

func (s *Store) UpsertEquipment(ctx context.Context, e domain.Equipment) error {
	if e.Specs == nil {
		e.Specs = domain.EmptyJSONObject()
	}
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO equipment (id,name,kind,specs) VALUES ($1,$2,$3,$4)
		ON CONFLICT (id) DO UPDATE SET name=$2,kind=$3,specs=$4`, e.ID, e.Name, e.Kind, e.Specs)
	return err
}

func (s *Store) ListEquipment(ctx context.Context) ([]domain.Equipment, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,name,kind,specs,created_at FROM equipment ORDER BY kind,name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Equipment
	for rows.Next() {
		var e domain.Equipment
		if err := rows.Scan(&e.ID, &e.Name, &e.Kind, &e.Specs, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if out == nil {
		out = []domain.Equipment{}
	}
	return out, rows.Err()
}

func (s *Store) UpsertRig(ctx context.Context, r domain.Rig) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO rigs (id,name,notes) VALUES ($1,$2,$3) ON CONFLICT (id) DO UPDATE SET name=$2,notes=$3`, r.ID, r.Name, r.Notes); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM rig_components WHERE rig_id=$1`, r.ID); err != nil {
		return err
	}
	for _, c := range r.Components {
		if _, err := tx.ExecContext(ctx, `INSERT INTO rig_components (rig_id,equipment_id,role) VALUES ($1,$2,$3)`, r.ID, c.EquipmentID, c.Role); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListRigs(ctx context.Context) ([]domain.Rig, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,name,notes,created_at FROM rigs ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Rig
	for rows.Next() {
		var r domain.Rig
		if err := rows.Scan(&r.ID, &r.Name, &r.Notes, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	for i := range out {
		cs, err := s.rigComps(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Components = cs
	}
	if out == nil {
		out = []domain.Rig{}
	}
	return out, nil
}

func (s *Store) rigComps(ctx context.Context, id uuid.UUID) ([]domain.RigComp, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT rc.equipment_id, rc.role, e.name, e.kind
		FROM rig_components rc JOIN equipment e ON e.id=rc.equipment_id WHERE rc.rig_id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RigComp
	for rows.Next() {
		var c domain.RigComp
		if err := rows.Scan(&c.EquipmentID, &c.Role, &c.Name, &c.Kind); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if out == nil {
		out = []domain.RigComp{}
	}
	return out, rows.Err()
}

func (s *Store) EnsureProfile(ctx context.Context, id uuid.UUID, name, ver string, weights, seeing []byte) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO weight_profiles (id,name,engine_ver,weights,seeing,notes)
		VALUES ($1,$2,$3,$4,$5,'default EMPIRICAL')
		ON CONFLICT (id) DO NOTHING`, id, name, ver, weights, seeing)
	return err
}

func (s *Store) InsertWeather(ctx context.Context, site uuid.UUID, provider string, from, to time.Time, payload []byte, hit bool) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO weather_snapshots (id,site_id,provider,fetched_at,valid_from,valid_to,payload,cache_hit)
		VALUES ($1,$2,$3,NOW(),$4,$5,$6,$7)`, uuid.New(), site, provider, from, to, payload, hit)
	return err
}

func (s *Store) ReplaceScores(ctx context.Context, slots []domain.ScoreSlot) error {
	if len(slots) == 0 {
		return nil
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM score_slots WHERE site_id=$1 AND target_id=$2`, slots[0].SiteID, slots[0].TargetID); err != nil {
		return err
	}
	for _, sl := range slots {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO score_slots (id,site_id,target_id,slot_utc,score,tier,factor_c,factor_s,factor_m,factor_a,factor_t,factor_l,factor_n,seeing_arcsec,seeing_derived,gate_reason,limiting_factor,engine_version,weight_profile_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
			sl.ID, sl.SiteID, sl.TargetID, sl.SlotUTC, sl.Score, sl.Tier,
			sl.FactorC, sl.FactorS, sl.FactorM, sl.FactorA, sl.FactorT, sl.FactorL, sl.FactorN,
			sl.SeeingArcsec, sl.SeeingDerived, sl.GateReason, sl.LimitingFactor, sl.EngineVersion, sl.WeightProfileID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListScores(ctx context.Context, site, target uuid.UUID, from, to time.Time) ([]domain.ScoreSlot, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id,site_id,target_id,slot_utc,score,tier,factor_c,factor_s,factor_m,factor_a,factor_t,factor_l,factor_n,seeing_arcsec,seeing_derived,gate_reason,limiting_factor,engine_version,weight_profile_id
		FROM score_slots WHERE site_id=$1 AND ($2 = '00000000-0000-0000-0000-000000000000'::uuid OR target_id=$2) AND slot_utc>=$3 AND slot_utc<$4
		ORDER BY slot_utc`, site, target, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ScoreSlot
	for rows.Next() {
		var sl domain.ScoreSlot
		if err := rows.Scan(&sl.ID, &sl.SiteID, &sl.TargetID, &sl.SlotUTC, &sl.Score, &sl.Tier,
			&sl.FactorC, &sl.FactorS, &sl.FactorM, &sl.FactorA, &sl.FactorT, &sl.FactorL, &sl.FactorN,
			&sl.SeeingArcsec, &sl.SeeingDerived, &sl.GateReason, &sl.LimitingFactor, &sl.EngineVersion, &sl.WeightProfileID); err != nil {
			return nil, err
		}
		out = append(out, sl)
	}
	if out == nil {
		out = []domain.ScoreSlot{}
	}
	return out, rows.Err()
}

func (s *Store) ReplaceWindows(ctx context.Context, site, target uuid.UUID, ws []domain.GoldenWindow) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM golden_windows WHERE site_id=$1 AND target_id=$2`, site, target); err != nil {
		return err
	}
	for _, w := range ws {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO golden_windows (id,site_id,target_id,start_utc,end_utc,start_local,end_local,tier,mean_score,peak_score,quality_integral,limiting_factor,engine_version,weight_profile_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
			w.ID, w.SiteID, w.TargetID, w.StartUTC, w.EndUTC, w.StartLocal, w.EndLocal, w.Tier, w.MeanScore, w.PeakScore, w.QualityIntegral, w.LimitingFactor, w.EngineVersion, w.WeightProfileID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListWindows(ctx context.Context, site uuid.UUID) ([]domain.GoldenWindow, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id,site_id,target_id,start_utc,end_utc,start_local,end_local,tier,mean_score,peak_score,quality_integral,limiting_factor,engine_version,weight_profile_id
		FROM golden_windows WHERE site_id=$1 ORDER BY quality_integral DESC`, site)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.GoldenWindow
	for rows.Next() {
		var w domain.GoldenWindow
		if err := rows.Scan(&w.ID, &w.SiteID, &w.TargetID, &w.StartUTC, &w.EndUTC, &w.StartLocal, &w.EndLocal, &w.Tier, &w.MeanScore, &w.PeakScore, &w.QualityIntegral, &w.LimitingFactor, &w.EngineVersion, &w.WeightProfileID); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	if out == nil {
		out = []domain.GoldenWindow{}
	}
	return out, rows.Err()
}

func (s *Store) UpsertPlan(ctx context.Context, p domain.Plan) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO plans (id,site_id,title,notes) VALUES ($1,$2,$3,$4) ON CONFLICT (id) DO UPDATE SET title=$3,notes=$4`, p.ID, p.SiteID, p.Title, p.Notes)
	return err
}

func (s *Store) AddPlanItem(ctx context.Context, it domain.PlanItem) error {
	if it.FilterSeq == nil {
		it.FilterSeq = domain.EmptyJSONArray()
	}
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO plan_items (id,plan_id,target_id,rig_id,start_utc,end_utc,exposure_s,frame_count,filter_seq,narrowband)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, it.ID, it.PlanID, it.TargetID, it.RigID, it.StartUTC, it.EndUTC, it.ExposureS, it.FrameCount, it.FilterSeq, it.Narrowband)
	return err
}

func (s *Store) ListPlans(ctx context.Context, site uuid.UUID) ([]domain.Plan, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,site_id,title,notes,created_at FROM plans WHERE $1 = '00000000-0000-0000-0000-000000000000'::uuid OR site_id=$1 ORDER BY created_at DESC`, site)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Plan
	for rows.Next() {
		var p domain.Plan
		if err := rows.Scan(&p.ID, &p.SiteID, &p.Title, &p.Notes, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	for i := range out {
		items, err := s.planItems(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Items = items
	}
	if out == nil {
		out = []domain.Plan{}
	}
	return out, nil
}

func (s *Store) planItems(ctx context.Context, id uuid.UUID) ([]domain.PlanItem, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,plan_id,target_id,rig_id,start_utc,end_utc,exposure_s,frame_count,filter_seq,narrowband,created_at FROM plan_items WHERE plan_id=$1 ORDER BY start_utc`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.PlanItem
	for rows.Next() {
		var it domain.PlanItem
		if err := rows.Scan(&it.ID, &it.PlanID, &it.TargetID, &it.RigID, &it.StartUTC, &it.EndUTC, &it.ExposureS, &it.FrameCount, &it.FilterSeq, &it.Narrowband, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	if out == nil {
		out = []domain.PlanItem{}
	}
	return out, rows.Err()
}

func (s *Store) SaveSession(ctx context.Context, sess domain.Session) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO sessions (id,rig_id,plan_item_id,state,progress_k,progress_n,remain_sec,last_error,source_mode,started_at,ended_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW())
		ON CONFLICT (id) DO UPDATE SET state=$4,progress_k=$5,progress_n=$6,remain_sec=$7,last_error=$8,ended_at=$11,updated_at=NOW()`,
		sess.ID, sess.RigID, sess.PlanItemID, sess.State, sess.ProgressK, sess.ProgressN, sess.RemainSec, sess.LastError, sess.SourceMode, sess.StartedAt, sess.EndedAt)
	return err
}

func (s *Store) AppendEvent(ctx context.Context, ev domain.SessionEvent) error {
	if ev.Context == nil {
		ev.Context = domain.EmptyJSONObject()
	}
	_, err := s.DB.ExecContext(ctx, `INSERT INTO session_events (id,session_id,from_state,to_state,class,context) VALUES ($1,$2,$3,$4,$5,$6)`,
		ev.ID, ev.SessionID, ev.FromState, ev.ToState, ev.Class, ev.Context)
	return err
}

func (s *Store) SaveCommand(ctx context.Context, commandID, sessionID uuid.UUID, verb string, payload, result []byte) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO session_commands (command_id,session_id,verb,payload,result) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (command_id) DO NOTHING`,
		commandID, sessionID, verb, payload, result)
	return err
}

func (s *Store) GetCommand(ctx context.Context, commandID uuid.UUID) (bool, []byte, error) {
	var raw []byte
	err := s.DB.QueryRowContext(ctx, `SELECT result FROM session_commands WHERE command_id=$1`, commandID).Scan(&raw)
	if err == sql.ErrNoRows {
		return false, nil, nil
	}
	return err == nil, raw, err
}

func (s *Store) AddExposure(ctx context.Context, sessionID uuid.UUID, seq int, filter string, dur float64, status, filename string) error {
	_, err := s.DB.ExecContext(ctx, `INSERT INTO exposures (id,session_id,seq,filter_name,duration_s,status,filename) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		uuid.New(), sessionID, seq, filter, dur, status, filename)
	return err
}

func (s *Store) IncompleteSessions(ctx context.Context) ([]domain.Session, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id,rig_id,plan_item_id,state,progress_k,progress_n,remain_sec,last_error,source_mode,started_at,ended_at,updated_at
		FROM sessions WHERE state NOT IN ('COMPLETED','ERROR','ABORTED','PARKED')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Session
	for rows.Next() {
		var x domain.Session
		if err := rows.Scan(&x.ID, &x.RigID, &x.PlanItemID, &x.State, &x.ProgressK, &x.ProgressN, &x.RemainSec, &x.LastError, &x.SourceMode, &x.StartedAt, &x.EndedAt, &x.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	if out == nil {
		out = []domain.Session{}
	}
	return out, rows.Err()
}

func (s *Store) ListSessions(ctx context.Context) ([]domain.Session, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,rig_id,plan_item_id,state,progress_k,progress_n,remain_sec,last_error,source_mode,started_at,ended_at,updated_at FROM sessions ORDER BY updated_at DESC LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Session
	for rows.Next() {
		var x domain.Session
		if err := rows.Scan(&x.ID, &x.RigID, &x.PlanItemID, &x.State, &x.ProgressK, &x.ProgressN, &x.RemainSec, &x.LastError, &x.SourceMode, &x.StartedAt, &x.EndedAt, &x.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	if out == nil {
		out = []domain.Session{}
	}
	return out, rows.Err()
}

func (s *Store) LogAPI(ctx context.Context, provider, endpoint string, latency time.Duration, code int, hit bool) {
	_, _ = s.DB.ExecContext(ctx, `INSERT INTO api_call_log (id,provider,endpoint,latency_ms,http_code,cache_hit,quota_used) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		uuid.New(), provider, endpoint, int(latency.Milliseconds()), code, hit, 1)
}

func (s *Store) InsertAlert(ctx context.Context, kind, msg string, site, sess *uuid.UUID) {
	_, _ = s.DB.ExecContext(ctx, `INSERT INTO alerts (id,site_id,session_id,kind,message) VALUES ($1,$2,$3,$4,$5)`, uuid.New(), site, sess, kind, msg)
}

func (s *Store) ListAlerts(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,kind,message,created_at FROM alerts ORDER BY created_at DESC LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id uuid.UUID
		var kind, msg string
		var at time.Time
		if err := rows.Scan(&id, &kind, &msg, &at); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "kind": kind, "message": msg, "created_at": at})
	}
	return out, rows.Err()
}

func MustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
