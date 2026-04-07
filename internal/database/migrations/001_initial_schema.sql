-- 001_initial_schema.sql
-- PodGen Network initial schema

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Shows
CREATE TABLE IF NOT EXISTS shows (
    id                       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug                     TEXT NOT NULL UNIQUE,
    name                     TEXT NOT NULL,
    description              TEXT NOT NULL DEFAULT '',
    language                 TEXT NOT NULL DEFAULT 'en',
    niche                    TEXT NOT NULL DEFAULT '',
    tone                     TEXT NOT NULL DEFAULT '',
    status                   TEXT NOT NULL DEFAULT 'active',
    cadence                  TEXT NOT NULL DEFAULT '',
    default_duration_minutes INTEGER NOT NULL DEFAULT 30,
    default_format           TEXT NOT NULL DEFAULT 'solo',
    intro_asset_id           UUID,
    outro_asset_id           UUID,
    cover_asset_id           UUID,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Host profiles
CREATE TABLE IF NOT EXISTS host_profiles (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    show_id               UUID NOT NULL REFERENCES shows(id) ON DELETE CASCADE,
    display_name          TEXT NOT NULL,
    persona_summary       TEXT NOT NULL DEFAULT '',
    speaking_style        TEXT NOT NULL DEFAULT '',
    provider              TEXT NOT NULL DEFAULT '',
    voice_id              TEXT NOT NULL DEFAULT '',
    speaking_rate         FLOAT NOT NULL DEFAULT 1.0,
    pronunciation_rules_json JSONB NOT NULL DEFAULT '{}',
    prompt_rules_json     JSONB NOT NULL DEFAULT '{}',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Episodes
CREATE TABLE IF NOT EXISTS episodes (
    id                       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    show_id                  UUID NOT NULL REFERENCES shows(id) ON DELETE CASCADE,
    status                   TEXT NOT NULL DEFAULT 'draft',
    topic                    TEXT NOT NULL DEFAULT '',
    angle                    TEXT NOT NULL DEFAULT '',
    title                    TEXT NOT NULL DEFAULT '',
    subtitle                 TEXT NOT NULL DEFAULT '',
    description              TEXT NOT NULL DEFAULT '',
    transcript               TEXT NOT NULL DEFAULT '',
    target_duration_seconds  INTEGER NOT NULL DEFAULT 1800,
    planned_publish_at       TIMESTAMPTZ,
    published_at             TIMESTAMPTZ,
    error_message            TEXT NOT NULL DEFAULT '',
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Episode sources
CREATE TABLE IF NOT EXISTS episode_sources (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    episode_id           UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    source_type          TEXT NOT NULL DEFAULT 'url',
    source_url           TEXT NOT NULL DEFAULT '',
    source_title         TEXT NOT NULL DEFAULT '',
    source_author        TEXT NOT NULL DEFAULT '',
    source_published_at  TIMESTAMPTZ,
    extracted_text       TEXT NOT NULL DEFAULT '',
    summary              TEXT NOT NULL DEFAULT '',
    relevance_score      FLOAT NOT NULL DEFAULT 0,
    trust_score          FLOAT NOT NULL DEFAULT 0,
    selected             BOOLEAN NOT NULL DEFAULT false,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Episode briefs
CREATE TABLE IF NOT EXISTS episode_briefs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    episode_id      UUID NOT NULL UNIQUE REFERENCES episodes(id) ON DELETE CASCADE,
    audience        TEXT NOT NULL DEFAULT '',
    tone            TEXT NOT NULL DEFAULT '',
    angle           TEXT NOT NULL DEFAULT '',
    key_points_json JSONB NOT NULL DEFAULT '[]',
    claims_json     JSONB NOT NULL DEFAULT '[]',
    cta             TEXT NOT NULL DEFAULT '',
    opening_hook    TEXT NOT NULL DEFAULT '',
    constraints_json JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Script drafts
CREATE TABLE IF NOT EXISTS script_drafts (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    episode_id   UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    version      INTEGER NOT NULL DEFAULT 1,
    format       TEXT NOT NULL DEFAULT 'solo',
    outline_json JSONB NOT NULL DEFAULT '[]',
    sections_json JSONB NOT NULL DEFAULT '[]',
    full_text    TEXT NOT NULL DEFAULT '',
    spoken_text  TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'draft',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audio assets
CREATE TABLE IF NOT EXISTS audio_assets (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    episode_id       UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    asset_type       TEXT NOT NULL,
    storage_key      TEXT NOT NULL DEFAULT '',
    mime_type        TEXT NOT NULL DEFAULT '',
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    metadata_json    JSONB NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Job runs
CREATE TABLE IF NOT EXISTS job_runs (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    episode_id    UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    stage         TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    attempt       INTEGER NOT NULL DEFAULT 1,
    input_json    JSONB NOT NULL DEFAULT '{}',
    output_json   JSONB NOT NULL DEFAULT '{}',
    started_at    TIMESTAMPTZ,
    finished_at   TIMESTAMPTZ,
    error_message TEXT NOT NULL DEFAULT ''
);

-- Publish targets
CREATE TABLE IF NOT EXISTS publish_targets (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    show_id       UUID NOT NULL REFERENCES shows(id) ON DELETE CASCADE,
    platform      TEXT NOT NULL,
    config_json   JSONB NOT NULL DEFAULT '{}',
    enabled       BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Analytics snapshots
CREATE TABLE IF NOT EXISTS analytics_snapshots (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    episode_id  UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    platform    TEXT NOT NULL,
    snapshot_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    plays       INTEGER NOT NULL DEFAULT 0,
    downloads   INTEGER NOT NULL DEFAULT 0,
    likes       INTEGER NOT NULL DEFAULT 0,
    shares      INTEGER NOT NULL DEFAULT 0,
    raw_json    JSONB NOT NULL DEFAULT '{}'
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_episodes_show_id ON episodes(show_id);
CREATE INDEX IF NOT EXISTS idx_episodes_status ON episodes(status);
CREATE INDEX IF NOT EXISTS idx_episode_sources_episode_id ON episode_sources(episode_id);
CREATE INDEX IF NOT EXISTS idx_script_drafts_episode_id ON script_drafts(episode_id);
CREATE INDEX IF NOT EXISTS idx_audio_assets_episode_id ON audio_assets(episode_id);
CREATE INDEX IF NOT EXISTS idx_job_runs_episode_id ON job_runs(episode_id);
CREATE INDEX IF NOT EXISTS idx_job_runs_stage ON job_runs(stage);
