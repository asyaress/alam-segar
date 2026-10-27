package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
)

const SessionCookieName = "alam_segar_session"

type User struct {
	ID              int64
	Role            string
	Status          string
	Name            string
	Email           string
	PasswordHash    string
	EmailVerifiedAt sql.NullTime
}

type Session struct {
	ID        int64
	User      *User
	CSRFToken string
	ExpiresAt time.Time
}

type SessionStore struct {
	db           *sql.DB
	secureCookie bool
	hashKey      []byte
	trustedProxy []*net.IPNet
	now          func() time.Time
}

func NewSessionStore(db *sql.DB, secureCookie bool, hashKey string, trustedProxyCIDRs []string) *SessionStore {
	trustedProxy := make([]*net.IPNet, 0, len(trustedProxyCIDRs))
	for _, cidr := range trustedProxyCIDRs {
		if _, network, err := net.ParseCIDR(cidr); err == nil {
			trustedProxy = append(trustedProxy, network)
		}
	}
	return &SessionStore{
		db:           db,
		secureCookie: secureCookie,
		hashKey:      []byte(hashKey),
		trustedProxy: trustedProxy,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

func (s *SessionStore) Start(ctx context.Context, w http.ResponseWriter, userID *int64) (*Session, error) {
	rawToken, err := randomHex(32)
	if err != nil {
		return nil, err
	}
	csrfToken, err := randomHex(32)
	if err != nil {
		return nil, err
	}
	now := s.now()
	expires := now.Add(12 * time.Hour)
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO user_sessions (token_hash, csrf_token, user_id, expires_at, last_seen_at)
		VALUES (?, ?, ?, ?, ?)`, tokenHash(rawToken), csrfToken, userID, expires, now)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	s.setCookie(w, rawToken, expires)
	return &Session{ID: id, CSRFToken: csrfToken, ExpiresAt: expires}, nil
}

func (s *SessionStore) Load(ctx context.Context, r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return nil, sql.ErrNoRows
	}

	var session Session
	var userID sql.NullInt64
	var role, status, name, email, passwordHash sql.NullString
	var emailVerifiedAt sql.NullTime
	var lastSeenAt time.Time
	err = s.db.QueryRowContext(ctx, `
		SELECT s.id, u.id, s.csrf_token, s.expires_at, s.last_seen_at,
		       u.role, u.status, u.name, u.email, u.password_hash, u.email_verified_at
		FROM user_sessions s
		LEFT JOIN users u ON u.id=s.user_id AND u.deleted_at IS NULL
		WHERE s.token_hash = ? AND s.expires_at > UTC_TIMESTAMP(6)`, tokenHash(cookie.Value)).
		Scan(&session.ID, &userID, &session.CSRFToken, &session.ExpiresAt, &lastSeenAt,
			&role, &status, &name, &email, &passwordHash, &emailVerifiedAt)
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		session.User = &User{
			ID:              userID.Int64,
			Role:            role.String,
			Status:          status.String,
			Name:            name.String,
			Email:           email.String,
			PasswordHash:    passwordHash.String,
			EmailVerifiedAt: emailVerifiedAt,
		}
	}

	if s.now().Sub(lastSeenAt) >= 5*time.Minute {
		_, _ = s.db.ExecContext(ctx, `UPDATE user_sessions SET last_seen_at = UTC_TIMESTAMP(6) WHERE id = ?`, session.ID)
	}
	return &session, nil
}

func (s *SessionStore) RotateToUser(ctx context.Context, w http.ResponseWriter, old *Session, userID int64) (*Session, error) {
	if old != nil {
		if _, err := s.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE id = ?`, old.ID); err != nil {
			return nil, err
		}
	}
	return s.Start(ctx, w, &userID)
}

func (s *SessionStore) Destroy(ctx context.Context, w http.ResponseWriter, session *Session) error {
	if session != nil {
		if _, err := s.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE id = ?`, session.ID); err != nil {
			return err
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (s *SessionStore) ValidCSRF(session *Session, submitted string) bool {
	if session == nil || submitted == "" {
		return false
	}
	return hmac.Equal([]byte(session.CSRFToken), []byte(submitted))
}

func (s *SessionStore) HashIP(r *http.Request) string {
	mac := hmac.New(sha256.New, s.hashKey)
	_, _ = mac.Write([]byte(s.clientIP(r)))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *SessionStore) clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		host = strings.TrimSpace(r.RemoteAddr)
	}
	peer := net.ParseIP(host)
	if peer == nil || !s.isTrustedProxy(peer) {
		return host
	}
	forwarded := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
	parsed := make([]net.IP, 0, len(forwarded))
	for _, item := range forwarded {
		ip := net.ParseIP(strings.TrimSpace(item))
		if ip == nil {
			return host
		}
		parsed = append(parsed, ip)
	}
	for index := len(parsed) - 1; index >= 0; index-- {
		if !s.isTrustedProxy(parsed[index]) {
			return parsed[index].String()
		}
	}
	if len(parsed) > 0 {
		return parsed[0].String()
	}
	return host
}

func (s *SessionStore) isTrustedProxy(ip net.IP) bool {
	for _, network := range s.trustedProxy {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func (s *SessionStore) setCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		Secure:   s.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}

func randomHex(byteLength int) (string, error) {
	if byteLength < 16 {
		return "", errors.New("token length terlalu pendek")
	}
	buffer := make([]byte, byteLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func tokenHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
