// Package audit appends to the append-only shippin_audit.events log.
package audit

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Sink struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Sink { return &Sink{db: db} }

// Event is a single audited action.
type Event struct {
	Action     string
	TargetKind string
	TargetID   string
	Data       any
}

// Append writes an audit event. It never blocks on failure beyond returning err.
func (s *Sink) Append(ctx context.Context, e Event) error {
	var raw []byte
	if e.Data != nil {
		raw, _ = json.Marshal(e.Data)
	} else {
		raw = []byte("{}")
	}
	if s == nil || s.db == nil {
		return nil
	}
	_, err := s.db.Exec(ctx,
		`INSERT INTO shippin_audit.events (action, target_kind, target_id, data)
		 VALUES ($1,$2,$3,$4)`,
		e.Action, e.TargetKind, e.TargetID, raw)
	return err
}
