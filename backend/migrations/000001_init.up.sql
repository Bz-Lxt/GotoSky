-- GotoSky schema. Advisory lock is taken in Go on the same Conn before migrate.

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name  TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sites (
    id            UUID PRIMARY KEY,
    name          TEXT NOT NULL,
    latitude      DOUBLE PRECISION NOT NULL CHECK (latitude BETWEEN -90 AND 90),
    longitude     DOUBLE PRECISION NOT NULL CHECK (longitude BETWEEN -180 AND 180),
    elevation_m   DOUBLE PRECISION NOT NULL DEFAULT 0,
    timezone      TEXT NOT NULL DEFAULT 'Asia/Shanghai',
    bortle        INT NOT NULL CHECK (bortle BETWEEN 1 AND 9),
    sqm           DOUBLE PRECISION NOT NULL,
    min_altitude  DOUBLE PRECISION NOT NULL DEFAULT 20,
    horizon_mask  JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS targets (
    id           UUID PRIMARY KEY,
    catalog_id   TEXT NOT NULL UNIQUE,
    name         TEXT NOT NULL,
    name_zh      TEXT NOT NULL DEFAULT '',
    ra_hours     DOUBLE PRECISION NOT NULL,
    dec_deg      DOUBLE PRECISION NOT NULL,
    mag          DOUBLE PRECISION,
    kind         TEXT NOT NULL,
    size_arcmin  DOUBLE PRECISION,
    notes        TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_targets_name ON targets (name);
CREATE INDEX IF NOT EXISTS idx_targets_kind ON targets (kind);

CREATE TABLE IF NOT EXISTS equipment (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    kind       TEXT NOT NULL CHECK (kind IN ('MOUNT','OTA','CAMERA','FILTER_WHEEL','GUIDE_SCOPE','GUIDE_CAMERA')),
    specs      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rigs (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    notes      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rig_components (
    rig_id       UUID NOT NULL REFERENCES rigs(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE RESTRICT,
    role         TEXT NOT NULL,
    PRIMARY KEY (rig_id, role)
);

CREATE TABLE IF NOT EXISTS weight_profiles (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    engine_ver  TEXT NOT NULL,
    weights     JSONB NOT NULL,
    seeing      JSONB NOT NULL,
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS weather_snapshots (
    id           UUID PRIMARY KEY,
    site_id      UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    provider     TEXT NOT NULL,
    fetched_at   TIMESTAMPTZ NOT NULL,
    valid_from   TIMESTAMPTZ NOT NULL,
    valid_to     TIMESTAMPTZ NOT NULL,
    payload      JSONB NOT NULL,
    cache_hit    BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_weather_site_time ON weather_snapshots (site_id, fetched_at DESC);

CREATE TABLE IF NOT EXISTS score_slots (
    id               UUID PRIMARY KEY,
    site_id          UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    target_id        UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    slot_utc         TIMESTAMPTZ NOT NULL,
    score            INT NOT NULL,
    tier             TEXT NOT NULL CHECK (tier IN ('GOLD','SILVER','BRONZE','POOR','UNUSABLE')),
    factor_c         DOUBLE PRECISION NOT NULL,
    factor_s         DOUBLE PRECISION NOT NULL,
    factor_m         DOUBLE PRECISION NOT NULL,
    factor_a         DOUBLE PRECISION NOT NULL,
    factor_t         DOUBLE PRECISION NOT NULL,
    factor_l         DOUBLE PRECISION NOT NULL,
    factor_n         DOUBLE PRECISION NOT NULL,
    seeing_arcsec    DOUBLE PRECISION NOT NULL,
    seeing_derived   BOOLEAN NOT NULL DEFAULT TRUE,
    gate_reason      TEXT NOT NULL DEFAULT '',
    limiting_factor  TEXT NOT NULL DEFAULT '',
    engine_version   TEXT NOT NULL,
    weight_profile_id UUID NOT NULL REFERENCES weight_profiles(id),
    UNIQUE (site_id, target_id, slot_utc, engine_version, weight_profile_id)
);
CREATE INDEX IF NOT EXISTS idx_score_site_slot ON score_slots (site_id, slot_utc);

CREATE TABLE IF NOT EXISTS golden_windows (
    id               UUID PRIMARY KEY,
    site_id          UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    target_id        UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    start_utc        TIMESTAMPTZ NOT NULL,
    end_utc          TIMESTAMPTZ NOT NULL,
    start_local      TIMESTAMP NOT NULL,
    end_local        TIMESTAMP NOT NULL,
    tier             TEXT NOT NULL,
    mean_score       DOUBLE PRECISION NOT NULL,
    peak_score       INT NOT NULL,
    quality_integral DOUBLE PRECISION NOT NULL,
    limiting_factor  TEXT NOT NULL DEFAULT '',
    engine_version   TEXT NOT NULL,
    weight_profile_id UUID NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_windows_site ON golden_windows (site_id, start_utc);

CREATE TABLE IF NOT EXISTS plans (
    id         UUID PRIMARY KEY,
    site_id    UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    title      TEXT NOT NULL,
    notes      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS plan_items (
    id            UUID PRIMARY KEY,
    plan_id       UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    target_id     UUID NOT NULL REFERENCES targets(id),
    rig_id        UUID REFERENCES rigs(id),
    start_utc     TIMESTAMPTZ NOT NULL,
    end_utc       TIMESTAMPTZ NOT NULL,
    exposure_s    DOUBLE PRECISION NOT NULL CHECK (exposure_s > 0),
    frame_count   INT NOT NULL DEFAULT 1 CHECK (frame_count > 0),
    filter_seq    JSONB NOT NULL DEFAULT '[]'::jsonb,
    narrowband    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sessions (
    id           UUID PRIMARY KEY,
    rig_id       UUID NOT NULL REFERENCES rigs(id),
    plan_item_id UUID REFERENCES plan_items(id),
    state        TEXT NOT NULL,
    progress_k   INT NOT NULL DEFAULT 0,
    progress_n   INT NOT NULL DEFAULT 0,
    remain_sec   DOUBLE PRECISION NOT NULL DEFAULT 0,
    last_error   TEXT NOT NULL DEFAULT '',
    source_mode  TEXT NOT NULL DEFAULT 'SIMULATED',
    started_at   TIMESTAMPTZ,
    ended_at     TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS session_events (
    id          UUID PRIMARY KEY,
    session_id  UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    seq         BIGSERIAL,
    from_state  TEXT NOT NULL,
    to_state    TEXT NOT NULL,
    class       TEXT NOT NULL DEFAULT '',
    context     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_events_session ON session_events (session_id, seq);

CREATE TABLE IF NOT EXISTS session_commands (
    command_id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    verb       TEXT NOT NULL,
    payload    JSONB NOT NULL DEFAULT '{}'::jsonb,
    result     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS exposures (
    id          UUID PRIMARY KEY,
    session_id  UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    seq         INT NOT NULL,
    filter_name TEXT NOT NULL DEFAULT 'L',
    duration_s  DOUBLE PRECISION NOT NULL,
    status      TEXT NOT NULL,
    filename    TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS alerts (
    id          UUID PRIMARY KEY,
    site_id     UUID REFERENCES sites(id),
    session_id  UUID REFERENCES sessions(id),
    kind        TEXT NOT NULL,
    message     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS api_call_log (
    id          UUID PRIMARY KEY,
    provider    TEXT NOT NULL,
    endpoint    TEXT NOT NULL,
    site_id     UUID,
    latency_ms  INT NOT NULL,
    http_code   INT NOT NULL,
    cache_hit   BOOLEAN NOT NULL DEFAULT FALSE,
    quota_used  INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_api_log_day ON api_call_log (provider, created_at);
