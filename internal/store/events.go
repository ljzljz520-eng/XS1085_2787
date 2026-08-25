package store

import (
	"context"
	"database/sql"
	"memorialcandle/internal/memorial"
)

func (s *SQLiteStore) SaveSpark(ctx context.Context, value memorial.Spark) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO sparks (id,candle_id,sequence,angle,radius,brightness,created_at) VALUES (?,?,?,?,?,?,?)`, value.ID, value.CandleID, value.Sequence, value.Angle, value.Radius, value.Brightness, stamp(value.CreatedAt))
	return err
}

func (s *SQLiteStore) ListSparks(ctx context.Context, candleID string, limit int) ([]memorial.Spark, error) {
	if limit < 1 {
		limit = 1
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,candle_id,sequence,angle,radius,brightness,created_at FROM sparks WHERE candle_id=? ORDER BY sequence DESC LIMIT ?`, candleID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []memorial.Spark{}
	for rows.Next() {
		var value memorial.Spark
		var created string
		if err := rows.Scan(&value.ID, &value.CandleID, &value.Sequence, &value.Angle, &value.Radius, &value.Brightness, &created); err != nil {
			return nil, err
		}
		value.CreatedAt, err = parseStamp(created)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (s *SQLiteStore) SaveReflection(ctx context.Context, value memorial.Reflection) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO reflections (id,memorial_id,visitor_id,text,moderated,created_at) VALUES (?,?,?,?,?,?)`, value.ID, value.MemorialID, value.VisitorID, value.Text, boolInt(value.Moderated), stamp(value.CreatedAt))
	return err
}

func (s *SQLiteStore) ListReflections(ctx context.Context, memorialID string) ([]memorial.Reflection, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,memorial_id,visitor_id,text,moderated,created_at FROM reflections WHERE memorial_id=? ORDER BY created_at ASC`, memorialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []memorial.Reflection{}
	for rows.Next() {
		var value memorial.Reflection
		var created string
		var moderated int
		if err := rows.Scan(&value.ID, &value.MemorialID, &value.VisitorID, &value.Text, &moderated, &created); err != nil {
			return nil, err
		}
		value.CreatedAt, err = parseStamp(created)
		if err != nil {
			return nil, err
		}
		value.Moderated = intBool(moderated)
		result = append(result, value)
	}
	return result, rows.Err()
}

func (s *SQLiteStore) SaveAudit(ctx context.Context, value memorial.AuditEntry) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO audits (id,memorial_id,action,subject,detail,created_at) VALUES (?,?,?,?,?,?)`, value.ID, value.MemorialID, value.Action, value.Subject, value.Detail, stamp(value.CreatedAt))
	return err
}

func (s *SQLiteStore) CountVisitors(ctx context.Context, memorialID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions WHERE memorial_id=? AND active=1`, memorialID).Scan(&count)
	return count, err
}

func (s *SQLiteStore) Ping(ctx context.Context) error {
	return s.db.QueryRowContext(ctx, `SELECT 1`).Scan(new(int))
}

func isNoRows(err error) bool { return err == sql.ErrNoRows }
