// Package account is the customer-facing Clyffe surface for WardenClyffe.
//
// Boundary: Clyffe is the customer plane. This store reads ONLY customer-safe
// data — account identity (shippin_core.tenants) and the customer's own orders
// (clyffe_core.orders). It must never read operator-only schemas
// (shippin_infra, shippin_audit) or surface secrets; per the Clyffe boundary
// rule, operator data is reachable only through sanitized Warden endpoints.
package account

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ db *pgxpool.Pool }

func NewStore(db *pgxpool.Pool) *Store { return &Store{db: db} }

// Account is the customer-safe projection of a tenant.
type Account struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Order is a customer-safe projection of a placed order.
type Order struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Product   string    `json:"product"`
	Tier      string    `json:"tier"`
	Repo      *string   `json:"repo"`
	Region    *string   `json:"region"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

var ErrUnauthenticated = errors.New("unauthenticated")
var ErrTenantNotFound = errors.New("organization not found")

// SessionSubject resolves the existing shippin-auth cookie without trusting
// caller-supplied identity headers. Raw tokens are never stored or logged.
func (s *Store) SessionSubject(ctx context.Context, token string) (string, error) {
	if len(token) != 64 {
		return "", ErrUnauthenticated
	}
	if _, err := hex.DecodeString(token); err != nil {
		return "", ErrUnauthenticated
	}
	hash := sha256.Sum256([]byte(token))
	var subject string
	err := s.db.QueryRow(ctx, `
		SELECT s.id::text FROM identity.sessions sess
		JOIN shippin_core.subjects s ON s.id = sess.subject_id
		WHERE sess.token_hash = $1 AND sess.expires_at > now()
		  AND s.kind = 'customer' AND s.is_active AND s.status = 'active'
		  AND s.email_verified_at IS NOT NULL`, hex.EncodeToString(hash[:])).Scan(&subject)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUnauthenticated
	}
	return subject, err
}

// ListAccounts returns only organizations the verified subject belongs to.
func (s *Store) ListAccounts(ctx context.Context, subject string) ([]Account, error) {
	rows, err := s.db.Query(ctx,
		`SELECT t.id, t.slug, t.name, t.created_at FROM shippin_core.tenants t
		 WHERE EXISTS (SELECT 1 FROM shippin_core.subject_tenant m
		               WHERE m.tenant_id = t.id AND m.subject_id = $1::uuid)
		 ORDER BY t.created_at, t.id`, subject)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Account{}
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.Slug, &a.Name, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ListOrders checks membership before setting an RLS context local to this
// read-only transaction. No pooled connection retains a customer's context.
func (s *Store) ListOrders(ctx context.Context, subject, tenantID string) ([]Order, error) {
	var tenant pgtype.UUID
	if err := tenant.Scan(tenantID); err != nil || !tenant.Valid {
		return nil, ErrTenantNotFound
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var member bool
	err = tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM shippin_core.subject_tenant
		WHERE subject_id = $1::uuid AND tenant_id = $2::uuid)`, subject, tenant).Scan(&member)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, ErrTenantNotFound
	}
	if _, err = tx.Exec(ctx, `SELECT set_config('app.current_tenant', $1, true)`, tenantID); err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `SELECT o.id, o.tenant_id, o.product, o.tier, o.repo, o.region,
		o.status::text, o.created_at FROM clyffe_core.orders o
		WHERE o.tenant_id = $1::uuid AND EXISTS (SELECT 1 FROM shippin_core.subject_tenant m
		  WHERE m.tenant_id = o.tenant_id AND m.subject_id = $2::uuid)
		ORDER BY o.created_at DESC, o.id`, tenant, subject)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Order{}
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.TenantID, &o.Product, &o.Tier, &o.Repo, &o.Region, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}
