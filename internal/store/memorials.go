package store

import (
	"context"
	"database/sql"
	"memorialcandle/internal/memorial"
)

func (s *SQLiteStore) SaveMemorial(ctx context.Context, value memorial.Memorial) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO memorials (id,title,dedication,created_at,candle_count,quiet) VALUES (?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET title=excluded.title,dedication=excluded.dedication,candle_count=excluded.candle_count,quiet=excluded.quiet`, value.ID, value.Title, value.Dedication, stamp(value.CreatedAt), value.CandleCount, boolInt(value.Quiet))
	return err
}

func (s *SQLiteStore) GetMemorial(ctx context.Context, id string) (memorial.Memorial, error) {
	var value memorial.Memorial
	var created string
	var quiet int
	err := s.db.QueryRowContext(ctx, `SELECT id,title,dedication,created_at,candle_count,quiet FROM memorials WHERE id=?`, id).Scan(&value.ID, &value.Title, &value.Dedication, &created, &value.CandleCount, &quiet)
	if err == sql.ErrNoRows {
		return memorial.Memorial{}, missing("memorial", id)
	}
	if err != nil {
		return value, err
	}
	value.CreatedAt, err = parseStamp(created)
	value.Quiet = intBool(quiet)
	return value, err
}

func (s *SQLiteStore) SaveCandle(ctx context.Context, value memorial.Candle) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO candles (id,memorial_id,visitor_id,message,color,intensity,lit_at,extinguished) VALUES (?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET message=excluded.message,color=excluded.color,intensity=excluded.intensity,extinguished=excluded.extinguished`, value.ID, value.MemorialID, value.VisitorID, value.Message, value.Color, value.Intensity, stamp(value.LitAt), boolInt(value.Extinguished))
	return err
}

func (s *SQLiteStore) GetCandle(ctx context.Context, id string) (memorial.Candle, error) {
	var value memorial.Candle
	var lit string
	var extinguished int
	err := s.db.QueryRowContext(ctx, `SELECT id,memorial_id,visitor_id,message,color,intensity,lit_at,extinguished FROM candles WHERE id=?`, id).Scan(&value.ID, &value.MemorialID, &value.VisitorID, &value.Message, &value.Color, &value.Intensity, &lit, &extinguished)
	if err == sql.ErrNoRows {
		return memorial.Candle{}, missing("candle", id)
	}
	if err != nil {
		return value, err
	}
	value.LitAt, err = parseStamp(lit)
	value.Extinguished = intBool(extinguished)
	return value, err
}

func (s *SQLiteStore) SaveSession(ctx context.Context, value memorial.VisitorSession) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO sessions (id,memorial_id,started_at,last_seen_at,angle,zoom,active) VALUES (?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET last_seen_at=excluded.last_seen_at,angle=excluded.angle,zoom=excluded.zoom,active=excluded.active`, value.ID, value.MemorialID, stamp(value.StartedAt), stamp(value.LastSeenAt), value.Angle, value.Zoom, boolInt(value.Active))
	return err
}

func (s *SQLiteStore) GetSession(ctx context.Context, id string) (memorial.VisitorSession, error) {
	var value memorial.VisitorSession
	var started, seen string
	var active int
	err := s.db.QueryRowContext(ctx, `SELECT id,memorial_id,started_at,last_seen_at,angle,zoom,active FROM sessions WHERE id=?`, id).Scan(&value.ID, &value.MemorialID, &started, &seen, &value.Angle, &value.Zoom, &active)
	if err == sql.ErrNoRows {
		return memorial.VisitorSession{}, missing("session", id)
	}
	if err != nil {
		return value, err
	}
	value.StartedAt, err = parseStamp(started)
	if err != nil {
		return value, err
	}
	value.LastSeenAt, err = parseStamp(seen)
	value.Active = intBool(active)
	return value, err
}
