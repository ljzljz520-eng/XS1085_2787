package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"memorialcandle/internal/memorial"
)

type AuditFilter struct {
	MemorialID string
	Action     string
	Subject    string
	Limit      int
}

type StoreCounts struct {
	Memorials   int `json:"memorials"`
	Candles     int `json:"candles"`
	Sparks      int `json:"sparks"`
	Sessions    int `json:"sessions"`
	Reflections int `json:"reflections"`
	Audits      int `json:"audits"`
}

func (s *SQLiteStore) SaveAnimationJob(ctx context.Context, job memorial.AnimationJob) error {
	if !job.Valid() {
		return errors.New("invalid animation job")
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO animation_jobs (id,candle_id,status,generated,started_at,stopped_at) VALUES (?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET status=excluded.status,generated=excluded.generated,stopped_at=excluded.stopped_at`, job.ID, job.CandleID, string(job.Status), job.Generated, stamp(job.StartedAt), stamp(job.StoppedAt))
	return err
}

func (s *SQLiteStore) GetAnimationJob(ctx context.Context, id string) (memorial.AnimationJob, error) {
	var job memorial.AnimationJob
	var status, started, stopped string
	err := s.db.QueryRowContext(ctx, `SELECT id,candle_id,status,generated,started_at,stopped_at FROM animation_jobs WHERE id=?`, id).Scan(&job.ID, &job.CandleID, &status, &job.Generated, &started, &stopped)
	if err == sql.ErrNoRows {
		return job, missing("animation job", id)
	}
	if err != nil {
		return job, err
	}
	job.Status = memorial.AnimationStatus(status)
	if job.StartedAt, err = parseStamp(started); err != nil {
		return job, err
	}
	if stopped != "" {
		job.StoppedAt, err = parseStamp(stopped)
	}
	return job, err
}

func (s *SQLiteStore) ListAudits(ctx context.Context, filter AuditFilter) ([]memorial.AuditEntry, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := `SELECT id,memorial_id,action,subject,detail,created_at FROM audits WHERE 1=1`
	args := []any{}
	if strings.TrimSpace(filter.MemorialID) != "" {
		query += ` AND memorial_id=?`
		args = append(args, filter.MemorialID)
	}
	if strings.TrimSpace(filter.Action) != "" {
		query += ` AND action=?`
		args = append(args, filter.Action)
	}
	if strings.TrimSpace(filter.Subject) != "" {
		query += ` AND subject=?`
		args = append(args, filter.Subject)
	}
	query += ` ORDER BY created_at ASC, id ASC LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entries := make([]memorial.AuditEntry, 0, limit)
	for rows.Next() {
		var entry memorial.AuditEntry
		var created string
		if err := rows.Scan(&entry.ID, &entry.MemorialID, &entry.Action, &entry.Subject, &entry.Detail, &created); err != nil {
			return nil, err
		}
		entry.CreatedAt, err = parseStamp(created)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (s *SQLiteStore) ListAuditEntries(ctx context.Context, memorialID string, limit int) ([]memorial.AuditEntry, error) {
	return s.ListAudits(ctx, AuditFilter{MemorialID: memorialID, Limit: limit})
}

func (s *SQLiteStore) Count(ctx context.Context, table string) (int, error) {
	allowed := map[string]bool{"memorials": true, "candles": true, "sparks": true, "sessions": true, "reflections": true, "audits": true, "animation_jobs": true}
	if !allowed[table] {
		return 0, fmt.Errorf("table %q is not countable", table)
	}
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+table).Scan(&count)
	return count, err
}

func (s *SQLiteStore) Counts(ctx context.Context) (StoreCounts, error) {
	var result StoreCounts
	queries := []struct {
		name   string
		target *int
	}{
		{"memorials", &result.Memorials},
		{"candles", &result.Candles},
		{"sparks", &result.Sparks},
		{"sessions", &result.Sessions},
		{"reflections", &result.Reflections},
		{"audits", &result.Audits},
	}
	for _, query := range queries {
		value, err := s.Count(ctx, query.name)
		if err != nil {
			return result, err
		}
		*query.target = value
	}
	return result, nil
}

func (s *SQLiteStore) CountSparks(ctx context.Context, candleID string) (int, error) {
	if strings.TrimSpace(candleID) == "" {
		return 0, errors.New("candle id is required")
	}
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sparks WHERE candle_id=?`, candleID).Scan(&count)
	return count, err
}

func (s *SQLiteStore) DeleteSparks(ctx context.Context, candleID string) (int64, error) {
	if strings.TrimSpace(candleID) == "" {
		return 0, errors.New("candle id is required")
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM sparks WHERE candle_id=?`, candleID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *SQLiteStore) SearchReflections(ctx context.Context, memorialID, phrase string) ([]memorial.Reflection, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,memorial_id,visitor_id,text,moderated,created_at FROM reflections WHERE memorial_id=? AND text LIKE ? ORDER BY created_at ASC`, memorialID, "%"+phrase+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []memorial.Reflection{}
	for rows.Next() {
		var value memorial.Reflection
		var text, created string
		var moderated int
		if err := rows.Scan(&value.ID, &value.MemorialID, &value.VisitorID, &text, &moderated, &created); err != nil {
			return nil, err
		}
		value.Text = text
		value.Moderated = intBool(moderated)
		value.CreatedAt, err = parseStamp(created)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func SortAudits(entries []memorial.AuditEntry) []memorial.AuditEntry {
	result := append([]memorial.AuditEntry(nil), entries...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].ID < result[j].ID
		}
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result
}

func SummarizeAudits(entries []memorial.AuditEntry) map[string]int {
	result := map[string]int{}
	for _, entry := range entries {
		kind := strings.TrimSpace(entry.Action)
		if kind == "" {
			kind = "unknown"
		}
		result[kind]++
	}
	return result
}

func (s *SQLiteStore) BackupRows(ctx context.Context, memorialID string) ([]string, error) {
	entries, err := s.ListAudits(ctx, AuditFilter{MemorialID: memorialID, Limit: 500})
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		result = append(result, fmt.Sprintf("%s|%s|%s|%s", entry.CreatedAt.Format("15:04:05"), entry.Action, entry.Subject, entry.Detail))
	}
	return result, nil
}
