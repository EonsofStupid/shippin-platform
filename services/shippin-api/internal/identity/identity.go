// Package identity is the Warden operator authentication context.
//
// Dev-bootstrap now: a single env-configured operator credential issuing
// in-memory bearer tokens. OIDC is the production upgrade (per the DDD identity
// context); the handler/route surface stays stable across that swap.
package identity

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

// ErrInvalid is returned for a bad username/password.
var ErrInvalid = errors.New("invalid credentials")

// Operator is the customer/operator-safe identity projection.
type Operator struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type session struct {
	op  Operator
	exp time.Time
}

// Store authenticates the bootstrap operator and tracks bearer-token sessions.
type Store struct {
	user string
	pass string
	ttl  time.Duration
	mu   sync.RWMutex
	sess map[string]session
}

func NewStore(user, pass string) *Store {
	return &Store{user: user, pass: pass, ttl: 12 * time.Hour, sess: map[string]session{}}
}

// Login validates the credential (constant-time) and issues a token.
func (s *Store) Login(user, pass string) (string, Operator, error) {
	if strings.TrimSpace(s.user) == "" || strings.TrimSpace(s.pass) == "" {
		return "", Operator{}, ErrInvalid
	}
	uOK := subtle.ConstantTimeCompare([]byte(user), []byte(s.user)) == 1
	pOK := subtle.ConstantTimeCompare([]byte(pass), []byte(s.pass)) == 1
	if !uOK || !pOK {
		return "", Operator{}, ErrInvalid
	}
	op := Operator{Username: s.user, Role: "operator"}
	tok := newToken()
	s.mu.Lock()
	s.sess[tok] = session{op: op, exp: time.Now().Add(s.ttl)}
	s.mu.Unlock()
	return tok, op, nil
}

// Validate resolves a token to its operator, honoring expiry.
func (s *Store) Validate(tok string) (Operator, bool) {
	if tok == "" {
		return Operator{}, false
	}
	s.mu.RLock()
	se, ok := s.sess[tok]
	s.mu.RUnlock()
	if !ok || time.Now().After(se.exp) {
		return Operator{}, false
	}
	return se.op, true
}

// Logout revokes a token.
func (s *Store) Logout(tok string) {
	s.mu.Lock()
	delete(s.sess, tok)
	s.mu.Unlock()
}

func newToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
