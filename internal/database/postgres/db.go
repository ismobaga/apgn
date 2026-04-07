package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/ismobaga/apgn/internal/domain/asset"
	"github.com/ismobaga/apgn/internal/domain/brief"
	"github.com/ismobaga/apgn/internal/domain/episode"
	"github.com/ismobaga/apgn/internal/domain/job"
	"github.com/ismobaga/apgn/internal/domain/script"
	"github.com/ismobaga/apgn/internal/domain/show"
	"github.com/ismobaga/apgn/internal/domain/source"
)

// DB wraps sql.DB and implements all domain repositories.
type DB struct {
	db *sql.DB
}

func New(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

// -------------------------
// Show repository
// -------------------------

func (d *DB) CreateShow(s *show.Show) error {
	s.ID = uuid.New()
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	if s.Status == "" {
		s.Status = show.StatusActive
	}
	_, err := d.db.Exec(`
		INSERT INTO shows (id, slug, name, description, language, niche, tone, status,
		                   cadence, default_duration_minutes, default_format,
		                   intro_asset_id, outro_asset_id, cover_asset_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		s.ID, s.Slug, s.Name, s.Description, s.Language, s.Niche, s.Tone, s.Status,
		s.Cadence, s.DefaultDurationMinutes, s.DefaultFormat,
		s.IntroAssetID, s.OutroAssetID, s.CoverAssetID, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (d *DB) GetShow(id uuid.UUID) (*show.Show, error) {
	row := d.db.QueryRow(`SELECT id, slug, name, description, language, niche, tone, status,
		cadence, default_duration_minutes, default_format,
		intro_asset_id, outro_asset_id, cover_asset_id, created_at, updated_at
		FROM shows WHERE id = $1`, id)
	return scanShow(row)
}

func (d *DB) GetShowBySlug(slug string) (*show.Show, error) {
	row := d.db.QueryRow(`SELECT id, slug, name, description, language, niche, tone, status,
		cadence, default_duration_minutes, default_format,
		intro_asset_id, outro_asset_id, cover_asset_id, created_at, updated_at
		FROM shows WHERE slug = $1`, slug)
	return scanShow(row)
}

func (d *DB) ListShows() ([]*show.Show, error) {
	rows, err := d.db.Query(`SELECT id, slug, name, description, language, niche, tone, status,
		cadence, default_duration_minutes, default_format,
		intro_asset_id, outro_asset_id, cover_asset_id, created_at, updated_at
		FROM shows ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shows []*show.Show
	for rows.Next() {
		s, err := scanShow(rows)
		if err != nil {
			return nil, err
		}
		shows = append(shows, s)
	}
	return shows, rows.Err()
}

func (d *DB) UpdateShow(s *show.Show) error {
	s.UpdatedAt = time.Now()
	_, err := d.db.Exec(`
		UPDATE shows SET slug=$1, name=$2, description=$3, language=$4, niche=$5, tone=$6,
		status=$7, cadence=$8, default_duration_minutes=$9, default_format=$10,
		intro_asset_id=$11, outro_asset_id=$12, cover_asset_id=$13, updated_at=$14
		WHERE id=$15`,
		s.Slug, s.Name, s.Description, s.Language, s.Niche, s.Tone, s.Status,
		s.Cadence, s.DefaultDurationMinutes, s.DefaultFormat,
		s.IntroAssetID, s.OutroAssetID, s.CoverAssetID, s.UpdatedAt, s.ID,
	)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanShow(s scanner) (*show.Show, error) {
	var sh show.Show
	err := s.Scan(
		&sh.ID, &sh.Slug, &sh.Name, &sh.Description, &sh.Language, &sh.Niche, &sh.Tone,
		&sh.Status, &sh.Cadence, &sh.DefaultDurationMinutes, &sh.DefaultFormat,
		&sh.IntroAssetID, &sh.OutroAssetID, &sh.CoverAssetID, &sh.CreatedAt, &sh.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &sh, err
}

// -------------------------
// Host profile repository
// -------------------------

func (d *DB) CreateHostProfile(h *show.HostProfile) error {
	h.ID = uuid.New()
	h.CreatedAt = time.Now()
	h.UpdatedAt = time.Now()
	if h.PronunciationRules == nil {
		h.PronunciationRules = []byte("{}")
	}
	if h.PromptRules == nil {
		h.PromptRules = []byte("{}")
	}
	_, err := d.db.Exec(`
		INSERT INTO host_profiles (id, show_id, display_name, persona_summary, speaking_style,
		provider, voice_id, speaking_rate, pronunciation_rules_json, prompt_rules_json, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		h.ID, h.ShowID, h.DisplayName, h.PersonaSummary, h.SpeakingStyle,
		h.Provider, h.VoiceID, h.SpeakingRate, h.PronunciationRules, h.PromptRules,
		h.CreatedAt, h.UpdatedAt,
	)
	return err
}

func (d *DB) GetHostProfile(id uuid.UUID) (*show.HostProfile, error) {
	row := d.db.QueryRow(`SELECT id, show_id, display_name, persona_summary, speaking_style,
		provider, voice_id, speaking_rate, pronunciation_rules_json, prompt_rules_json, created_at, updated_at
		FROM host_profiles WHERE id = $1`, id)
	return scanHostProfile(row)
}

func (d *DB) ListHostProfiles(showID uuid.UUID) ([]*show.HostProfile, error) {
	rows, err := d.db.Query(`SELECT id, show_id, display_name, persona_summary, speaking_style,
		provider, voice_id, speaking_rate, pronunciation_rules_json, prompt_rules_json, created_at, updated_at
		FROM host_profiles WHERE show_id = $1 ORDER BY created_at ASC`, showID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*show.HostProfile
	for rows.Next() {
		h, err := scanHostProfile(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, h)
	}
	return profiles, rows.Err()
}

func scanHostProfile(s scanner) (*show.HostProfile, error) {
	var h show.HostProfile
	err := s.Scan(
		&h.ID, &h.ShowID, &h.DisplayName, &h.PersonaSummary, &h.SpeakingStyle,
		&h.Provider, &h.VoiceID, &h.SpeakingRate,
		&h.PronunciationRules, &h.PromptRules,
		&h.CreatedAt, &h.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &h, err
}

// -------------------------
// Episode repository
// -------------------------

func (d *DB) CreateEpisode(e *episode.Episode) error {
	e.ID = uuid.New()
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
	if e.Status == "" {
		e.Status = episode.StatusDraft
	}
	_, err := d.db.Exec(`
		INSERT INTO episodes (id, show_id, status, topic, angle, title, subtitle, description,
		transcript, target_duration_seconds, planned_publish_at, published_at, error_message, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		e.ID, e.ShowID, e.Status, e.Topic, e.Angle, e.Title, e.Subtitle, e.Description,
		e.Transcript, e.TargetDurationSeconds, e.PlannedPublishAt, e.PublishedAt,
		e.ErrorMessage, e.CreatedAt, e.UpdatedAt,
	)
	return err
}

func (d *DB) GetEpisode(id uuid.UUID) (*episode.Episode, error) {
	row := d.db.QueryRow(`SELECT id, show_id, status, topic, angle, title, subtitle, description,
		transcript, target_duration_seconds, planned_publish_at, published_at, error_message, created_at, updated_at
		FROM episodes WHERE id = $1`, id)
	return scanEpisode(row)
}

func (d *DB) ListEpisodes(showID *uuid.UUID, status *episode.Status) ([]*episode.Episode, error) {
	query := `SELECT id, show_id, status, topic, angle, title, subtitle, description,
		transcript, target_duration_seconds, planned_publish_at, published_at, error_message, created_at, updated_at
		FROM episodes WHERE 1=1`
	args := []any{}
	idx := 1
	if showID != nil {
		query += fmt.Sprintf(" AND show_id = $%d", idx)
		args = append(args, *showID)
		idx++
	}
	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, *status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eps []*episode.Episode
	for rows.Next() {
		e, err := scanEpisode(rows)
		if err != nil {
			return nil, err
		}
		eps = append(eps, e)
	}
	return eps, rows.Err()
}

func (d *DB) UpdateEpisode(e *episode.Episode) error {
	e.UpdatedAt = time.Now()
	_, err := d.db.Exec(`
		UPDATE episodes SET show_id=$1, status=$2, topic=$3, angle=$4, title=$5, subtitle=$6,
		description=$7, transcript=$8, target_duration_seconds=$9, planned_publish_at=$10,
		published_at=$11, error_message=$12, updated_at=$13
		WHERE id=$14`,
		e.ShowID, e.Status, e.Topic, e.Angle, e.Title, e.Subtitle,
		e.Description, e.Transcript, e.TargetDurationSeconds, e.PlannedPublishAt,
		e.PublishedAt, e.ErrorMessage, e.UpdatedAt, e.ID,
	)
	return err
}

func (d *DB) UpdateEpisodeStatus(id uuid.UUID, status episode.Status, errMsg string) error {
	_, err := d.db.Exec(
		`UPDATE episodes SET status=$1, error_message=$2, updated_at=$3 WHERE id=$4`,
		status, errMsg, time.Now(), id,
	)
	return err
}

func scanEpisode(s scanner) (*episode.Episode, error) {
	var e episode.Episode
	err := s.Scan(
		&e.ID, &e.ShowID, &e.Status, &e.Topic, &e.Angle, &e.Title, &e.Subtitle,
		&e.Description, &e.Transcript, &e.TargetDurationSeconds,
		&e.PlannedPublishAt, &e.PublishedAt, &e.ErrorMessage, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &e, err
}

// -------------------------
// Source repository
// -------------------------

func (d *DB) CreateSource(s *source.EpisodeSource) error {
	s.ID = uuid.New()
	s.CreatedAt = time.Now()
	_, err := d.db.Exec(`
		INSERT INTO episode_sources (id, episode_id, source_type, source_url, source_title, source_author,
		source_published_at, extracted_text, summary, relevance_score, trust_score, selected, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		s.ID, s.EpisodeID, s.SourceType, s.SourceURL, s.SourceTitle, s.SourceAuthor,
		s.SourcePublishedAt, s.ExtractedText, s.Summary, s.RelevanceScore,
		s.TrustScore, s.Selected, s.CreatedAt,
	)
	return err
}

func (d *DB) GetSource(id uuid.UUID) (*source.EpisodeSource, error) {
	row := d.db.QueryRow(`SELECT id, episode_id, source_type, source_url, source_title, source_author,
		source_published_at, extracted_text, summary, relevance_score, trust_score, selected, created_at
		FROM episode_sources WHERE id = $1`, id)
	return scanSource(row)
}

func (d *DB) ListSources(episodeID uuid.UUID) ([]*source.EpisodeSource, error) {
	rows, err := d.db.Query(`SELECT id, episode_id, source_type, source_url, source_title, source_author,
		source_published_at, extracted_text, summary, relevance_score, trust_score, selected, created_at
		FROM episode_sources WHERE episode_id = $1 ORDER BY created_at ASC`, episodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []*source.EpisodeSource
	for rows.Next() {
		s, err := scanSource(rows)
		if err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

func (d *DB) UpdateSource(s *source.EpisodeSource) error {
	_, err := d.db.Exec(`
		UPDATE episode_sources SET source_type=$1, source_url=$2, source_title=$3, source_author=$4,
		source_published_at=$5, extracted_text=$6, summary=$7, relevance_score=$8,
		trust_score=$9, selected=$10
		WHERE id=$11`,
		s.SourceType, s.SourceURL, s.SourceTitle, s.SourceAuthor,
		s.SourcePublishedAt, s.ExtractedText, s.Summary, s.RelevanceScore,
		s.TrustScore, s.Selected, s.ID,
	)
	return err
}

func scanSource(s scanner) (*source.EpisodeSource, error) {
	var src source.EpisodeSource
	err := s.Scan(
		&src.ID, &src.EpisodeID, &src.SourceType, &src.SourceURL, &src.SourceTitle,
		&src.SourceAuthor, &src.SourcePublishedAt, &src.ExtractedText, &src.Summary,
		&src.RelevanceScore, &src.TrustScore, &src.Selected, &src.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &src, err
}

// -------------------------
// Brief repository
// -------------------------

func (d *DB) CreateBrief(b *brief.EpisodeBrief) error {
	b.ID = uuid.New()
	b.CreatedAt = time.Now()
	b.UpdatedAt = time.Now()
	if b.KeyPoints == nil {
		b.KeyPoints = json.RawMessage("[]")
	}
	if b.Claims == nil {
		b.Claims = json.RawMessage("[]")
	}
	if b.Constraints == nil {
		b.Constraints = json.RawMessage("{}")
	}
	_, err := d.db.Exec(`
		INSERT INTO episode_briefs (id, episode_id, audience, tone, angle, key_points_json,
		claims_json, cta, opening_hook, constraints_json, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		b.ID, b.EpisodeID, b.Audience, b.Tone, b.Angle, b.KeyPoints,
		b.Claims, b.CTA, b.OpeningHook, b.Constraints, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (d *DB) GetBriefByEpisode(episodeID uuid.UUID) (*brief.EpisodeBrief, error) {
	row := d.db.QueryRow(`SELECT id, episode_id, audience, tone, angle, key_points_json,
		claims_json, cta, opening_hook, constraints_json, created_at, updated_at
		FROM episode_briefs WHERE episode_id = $1`, episodeID)
	return scanBrief(row)
}

func (d *DB) UpdateBrief(b *brief.EpisodeBrief) error {
	b.UpdatedAt = time.Now()
	_, err := d.db.Exec(`
		UPDATE episode_briefs SET audience=$1, tone=$2, angle=$3, key_points_json=$4,
		claims_json=$5, cta=$6, opening_hook=$7, constraints_json=$8, updated_at=$9
		WHERE id=$10`,
		b.Audience, b.Tone, b.Angle, b.KeyPoints, b.Claims,
		b.CTA, b.OpeningHook, b.Constraints, b.UpdatedAt, b.ID,
	)
	return err
}

func scanBrief(s scanner) (*brief.EpisodeBrief, error) {
	var b brief.EpisodeBrief
	err := s.Scan(
		&b.ID, &b.EpisodeID, &b.Audience, &b.Tone, &b.Angle, &b.KeyPoints,
		&b.Claims, &b.CTA, &b.OpeningHook, &b.Constraints, &b.CreatedAt, &b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &b, err
}

// -------------------------
// Script repository
// -------------------------

func (d *DB) CreateDraft(dr *script.ScriptDraft) error {
	dr.ID = uuid.New()
	dr.CreatedAt = time.Now()
	if dr.Status == "" {
		dr.Status = script.ScriptStatusDraft
	}
	if dr.Outline == nil {
		dr.Outline = json.RawMessage("[]")
	}
	if dr.Sections == nil {
		dr.Sections = json.RawMessage("[]")
	}
	// Get next version number
	var maxVersion sql.NullInt64
	_ = d.db.QueryRow(`SELECT MAX(version) FROM script_drafts WHERE episode_id = $1`, dr.EpisodeID).Scan(&maxVersion)
	if maxVersion.Valid {
		dr.Version = int(maxVersion.Int64) + 1
	} else {
		dr.Version = 1
	}
	_, err := d.db.Exec(`
		INSERT INTO script_drafts (id, episode_id, version, format, outline_json, sections_json,
		full_text, spoken_text, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		dr.ID, dr.EpisodeID, dr.Version, dr.Format, dr.Outline, dr.Sections,
		dr.FullText, dr.SpokenText, dr.Status, dr.CreatedAt,
	)
	return err
}

func (d *DB) GetDraft(id uuid.UUID) (*script.ScriptDraft, error) {
	row := d.db.QueryRow(`SELECT id, episode_id, version, format, outline_json, sections_json,
		full_text, spoken_text, status, created_at
		FROM script_drafts WHERE id = $1`, id)
	return scanDraft(row)
}

func (d *DB) GetLatestDraft(episodeID uuid.UUID) (*script.ScriptDraft, error) {
	row := d.db.QueryRow(`SELECT id, episode_id, version, format, outline_json, sections_json,
		full_text, spoken_text, status, created_at
		FROM script_drafts WHERE episode_id = $1 ORDER BY version DESC LIMIT 1`, episodeID)
	return scanDraft(row)
}

func (d *DB) ListDrafts(episodeID uuid.UUID) ([]*script.ScriptDraft, error) {
	rows, err := d.db.Query(`SELECT id, episode_id, version, format, outline_json, sections_json,
		full_text, spoken_text, status, created_at
		FROM script_drafts WHERE episode_id = $1 ORDER BY version DESC`, episodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []*script.ScriptDraft
	for rows.Next() {
		dr, err := scanDraft(rows)
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, dr)
	}
	return drafts, rows.Err()
}

func (d *DB) UpdateDraft(dr *script.ScriptDraft) error {
	_, err := d.db.Exec(`
		UPDATE script_drafts SET format=$1, outline_json=$2, sections_json=$3,
		full_text=$4, spoken_text=$5, status=$6
		WHERE id=$7`,
		dr.Format, dr.Outline, dr.Sections, dr.FullText, dr.SpokenText, dr.Status, dr.ID,
	)
	return err
}

func scanDraft(s scanner) (*script.ScriptDraft, error) {
	var d script.ScriptDraft
	err := s.Scan(
		&d.ID, &d.EpisodeID, &d.Version, &d.Format, &d.Outline, &d.Sections,
		&d.FullText, &d.SpokenText, &d.Status, &d.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &d, err
}

// -------------------------
// Asset repository
// -------------------------

func (d *DB) CreateAsset(a *asset.AudioAsset) error {
	a.ID = uuid.New()
	a.CreatedAt = time.Now()
	if a.Metadata == nil {
		a.Metadata = json.RawMessage("{}")
	}
	_, err := d.db.Exec(`
		INSERT INTO audio_assets (id, episode_id, asset_type, storage_key, mime_type,
		duration_seconds, metadata_json, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.EpisodeID, a.AssetType, a.StorageKey, a.MimeType,
		a.DurationSeconds, a.Metadata, a.CreatedAt,
	)
	return err
}

func (d *DB) GetAsset(id uuid.UUID) (*asset.AudioAsset, error) {
	row := d.db.QueryRow(`SELECT id, episode_id, asset_type, storage_key, mime_type,
		duration_seconds, metadata_json, created_at
		FROM audio_assets WHERE id = $1`, id)
	return scanAsset(row)
}

func (d *DB) ListAssets(episodeID uuid.UUID) ([]*asset.AudioAsset, error) {
	rows, err := d.db.Query(`SELECT id, episode_id, asset_type, storage_key, mime_type,
		duration_seconds, metadata_json, created_at
		FROM audio_assets WHERE episode_id = $1 ORDER BY created_at ASC`, episodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*asset.AudioAsset
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, rows.Err()
}

func scanAsset(s scanner) (*asset.AudioAsset, error) {
	var a asset.AudioAsset
	err := s.Scan(
		&a.ID, &a.EpisodeID, &a.AssetType, &a.StorageKey, &a.MimeType,
		&a.DurationSeconds, &a.Metadata, &a.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

// -------------------------
// Job repository
// -------------------------

func (d *DB) CreateJobRun(j *job.JobRun) error {
	j.ID = uuid.New()
	if j.Status == "" {
		j.Status = job.JobStatusPending
	}
	if j.Input == nil {
		j.Input = json.RawMessage("{}")
	}
	if j.Output == nil {
		j.Output = json.RawMessage("{}")
	}
	_, err := d.db.Exec(`
		INSERT INTO job_runs (id, episode_id, stage, status, attempt, input_json, output_json,
		started_at, finished_at, error_message)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		j.ID, j.EpisodeID, j.Stage, j.Status, j.Attempt, j.Input, j.Output,
		j.StartedAt, j.FinishedAt, j.ErrorMessage,
	)
	return err
}

func (d *DB) GetJobRun(id uuid.UUID) (*job.JobRun, error) {
	row := d.db.QueryRow(`SELECT id, episode_id, stage, status, attempt, input_json, output_json,
		started_at, finished_at, error_message
		FROM job_runs WHERE id = $1`, id)
	return scanJobRun(row)
}

func (d *DB) ListJobRuns(episodeID uuid.UUID) ([]*job.JobRun, error) {
	rows, err := d.db.Query(`SELECT id, episode_id, stage, status, attempt, input_json, output_json,
		started_at, finished_at, error_message
		FROM job_runs WHERE episode_id = $1 ORDER BY id ASC`, episodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*job.JobRun
	for rows.Next() {
		j, err := scanJobRun(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (d *DB) UpdateJobRun(j *job.JobRun) error {
	_, err := d.db.Exec(`
		UPDATE job_runs SET status=$1, attempt=$2, output_json=$3,
		started_at=$4, finished_at=$5, error_message=$6
		WHERE id=$7`,
		j.Status, j.Attempt, j.Output, j.StartedAt, j.FinishedAt, j.ErrorMessage, j.ID,
	)
	return err
}

func scanJobRun(s scanner) (*job.JobRun, error) {
	var j job.JobRun
	err := s.Scan(
		&j.ID, &j.EpisodeID, &j.Stage, &j.Status, &j.Attempt,
		&j.Input, &j.Output, &j.StartedAt, &j.FinishedAt, &j.ErrorMessage,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &j, err
}
