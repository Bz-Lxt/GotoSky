package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const EngineVersion = "sve-1.0.0"

type Site struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Latitude    float64         `json:"latitude"`
	Longitude   float64         `json:"longitude"`
	ElevationM  float64         `json:"elevation_m"`
	Timezone    string          `json:"timezone"`
	Bortle      int             `json:"bortle"`
	SQM         float64         `json:"sqm"`
	MinAltitude float64         `json:"min_altitude"`
	HorizonMask json.RawMessage `json:"horizon_mask"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type HorizonPoint struct {
	AzimuthDeg  float64 `json:"azimuth_deg"`
	AltitudeDeg float64 `json:"altitude_deg"`
}

type Target struct {
	ID         uuid.UUID `json:"id"`
	CatalogID  string    `json:"catalog_id"`
	Name       string    `json:"name"`
	NameZH     string    `json:"name_zh"`
	RAHours    float64   `json:"ra_hours"`
	DecDeg     float64   `json:"dec_deg"`
	Mag        *float64  `json:"mag"`
	Kind       string    `json:"kind"`
	SizeArcmin *float64  `json:"size_arcmin"`
	Notes      string    `json:"notes"`
}

type Equipment struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Kind      string          `json:"kind"`
	Specs     json.RawMessage `json:"specs"`
	CreatedAt time.Time       `json:"created_at"`
}

type Rig struct {
	ID         uuid.UUID   `json:"id"`
	Name       string      `json:"name"`
	Notes      string      `json:"notes"`
	Components []RigComp   `json:"components"`
	FOVArcmin  float64     `json:"fov_arcmin"`
	ScalePPS   float64     `json:"scale_arcsec_px"`
	CreatedAt  time.Time   `json:"created_at"`
}

type RigComp struct {
	EquipmentID uuid.UUID `json:"equipment_id"`
	Role        string    `json:"role"`
	Name        string    `json:"name,omitempty"`
	Kind        string    `json:"kind,omitempty"`
}

type WeatherHour struct {
	TimeUTC      time.Time
	CloudLow     float64
	CloudMid     float64
	CloudHigh    float64
	RH           float64
	TempC        float64
	DewC         float64
	VisibilityM  float64
	PrecipProb   float64
	Wind10MS     float64
	Gust10MS     float64
	Wind250MS    float64
	Wind500MS    float64
	Wind850MS    float64
}

type ScoreResult struct {
	Score           int     `json:"score"`
	Tier            string  `json:"tier"`
	C, S, M, A, T, L, N float64
	SeeingArcsec    float64 `json:"seeing_arcsec"`
	SeeingDerived   bool    `json:"seeing_derived"`
	GateReason      string  `json:"gate_reason"`
	LimitingFactor  string  `json:"limiting_factor"`
	EngineVersion   string  `json:"engine_version"`
}

type ScoreSlot struct {
	ID              uuid.UUID `json:"id"`
	SiteID          uuid.UUID `json:"site_id"`
	TargetID        uuid.UUID `json:"target_id"`
	SlotUTC         time.Time `json:"slot_utc"`
	Score           int       `json:"score"`
	Tier            string    `json:"tier"`
	FactorC         float64   `json:"factor_c"`
	FactorS         float64   `json:"factor_s"`
	FactorM         float64   `json:"factor_m"`
	FactorA         float64   `json:"factor_a"`
	FactorT         float64   `json:"factor_t"`
	FactorL         float64   `json:"factor_l"`
	FactorN         float64   `json:"factor_n"`
	SeeingArcsec    float64   `json:"seeing_arcsec"`
	SeeingDerived   bool      `json:"seeing_derived"`
	GateReason      string    `json:"gate_reason"`
	LimitingFactor  string    `json:"limiting_factor"`
	EngineVersion   string    `json:"engine_version"`
	WeightProfileID uuid.UUID `json:"weight_profile_id"`
}

type GoldenWindow struct {
	ID               uuid.UUID `json:"id"`
	SiteID           uuid.UUID `json:"site_id"`
	TargetID         uuid.UUID `json:"target_id"`
	StartUTC         time.Time `json:"start_utc"`
	EndUTC           time.Time `json:"end_utc"`
	StartLocal       time.Time `json:"start_local"`
	EndLocal         time.Time `json:"end_local"`
	Tier             string    `json:"tier"`
	MeanScore        float64   `json:"mean_score"`
	PeakScore        int       `json:"peak_score"`
	QualityIntegral  float64   `json:"quality_integral"`
	LimitingFactor   string    `json:"limiting_factor"`
	EngineVersion    string    `json:"engine_version"`
	WeightProfileID  uuid.UUID `json:"weight_profile_id"`
}

type Plan struct {
	ID        uuid.UUID  `json:"id"`
	SiteID    uuid.UUID  `json:"site_id"`
	Title     string     `json:"title"`
	Notes     string     `json:"notes"`
	Items     []PlanItem `json:"items"`
	CreatedAt time.Time  `json:"created_at"`
}

type PlanItem struct {
	ID          uuid.UUID       `json:"id"`
	PlanID      uuid.UUID       `json:"plan_id"`
	TargetID    uuid.UUID       `json:"target_id"`
	RigID       *uuid.UUID      `json:"rig_id"`
	StartUTC    time.Time       `json:"start_utc"`
	EndUTC      time.Time       `json:"end_utc"`
	ExposureS   float64         `json:"exposure_s"`
	FrameCount  int             `json:"frame_count"`
	FilterSeq   json.RawMessage `json:"filter_seq"`
	Narrowband  bool            `json:"narrowband"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Session struct {
	ID         uuid.UUID  `json:"id"`
	RigID      uuid.UUID  `json:"rig_id"`
	PlanItemID *uuid.UUID `json:"plan_item_id"`
	State      string     `json:"state"`
	ProgressK  int        `json:"progress_k"`
	ProgressN  int        `json:"progress_n"`
	RemainSec  float64    `json:"remain_sec"`
	LastError  string     `json:"last_error"`
	SourceMode string     `json:"source_mode"`
	StartedAt  *time.Time `json:"started_at"`
	EndedAt    *time.Time `json:"ended_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type SessionEvent struct {
	ID         uuid.UUID       `json:"id"`
	SessionID  uuid.UUID       `json:"session_id"`
	Seq        int64           `json:"seq"`
	FromState  string          `json:"from_state"`
	ToState    string          `json:"to_state"`
	Class      string          `json:"class"`
	Context    json.RawMessage `json:"context"`
	CreatedAt  time.Time       `json:"created_at"`
}

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name"`
}

type Telemetry struct {
	SessionID   uuid.UUID      `json:"session_id"`
	RigID       uuid.UUID      `json:"rig_id"`
	State       string         `json:"state"`
	ProgressK   int            `json:"progress_k"`
	ProgressN   int            `json:"progress_n"`
	RemainSec   float64        `json:"remain_sec"`
	Sensors     map[string]any `json:"sensors"`
	Source      string         `json:"source"`
	Alert       string         `json:"alert,omitempty"`
}

func EmptyJSONArray() json.RawMessage { return json.RawMessage("[]") }
func EmptyJSONObject() json.RawMessage { return json.RawMessage("{}") }
