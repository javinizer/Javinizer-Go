package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	credentialFilename = "auth.credentials.json"
	minPasswordLength  = 8
	sessionIDBytes     = 32
	saltLength         = 16
	maxActiveSessions  = 256

	defaultArgon2Memory  uint32 = 64 * 1024 // 64 MB
	defaultArgon2Time    uint32 = 1
	defaultArgon2Threads uint8  = 4
	defaultArgon2KeyLen  uint32 = 32
)

const (
	maxFailedLoginAttempts = 5
	failedLoginWindow      = 5 * time.Minute
	loginLockoutDuration   = 5 * time.Minute
)

// DefaultSessionTTL is the default authenticated session lifetime.
const DefaultSessionTTL = 24 * time.Hour

var (
	ErrAuthNotInitialized = errors.New("authentication is not initialized")
	ErrAuthAlreadySet     = errors.New("authentication is already initialized")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidSession     = errors.New("invalid or expired session")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrWeakPassword       = errors.New("weak password")
	ErrLoginRateLimited   = errors.New("too many login attempts")
)

type argon2Params struct {
	Memory  uint32
	Time    uint32
	Threads uint8
	KeyLen  uint32
}

type storedCredentials struct {
	Username string
	Salt     []byte
	Hash     []byte
	Params   argon2Params
}

type sessionRecord struct {
	Username  string
	ExpiresAt time.Time
}

type credentialFile struct {
	Version   int    `json:"version"`
	Username  string `json:"username"`
	Salt      string `json:"salt"`
	Hash      string `json:"hash"`
	Memory    uint32 `json:"memory"`
	Time      uint32 `json:"time"`
	Threads   uint8  `json:"threads"`
	KeyLen    uint32 `json:"key_len"`
	CreatedAt string `json:"created_at,omitempty"`
}

// AuthManager manages single-user credentials and in-memory sessions.
type AuthManager struct {
	mu             sync.RWMutex
	credentialPath string
	credentials    *storedCredentials
	sessions       map[string]sessionRecord
	sessionTTL     time.Duration
	nowFn          func() time.Time
	randReader     io.Reader

	failedLoginCount       int
	failedLoginWindowStart time.Time
	loginBlockedUntil      time.Time
}

// CredentialPathForConfig returns the auth credential file path next to config.
func CredentialPathForConfig(configFile string) string {
	return filepath.Join(filepath.Dir(configFile), credentialFilename)
}

// NewAuthManager creates an auth manager and loads credentials from disk if present.
func NewAuthManager(configFile string, sessionTTL time.Duration) (*AuthManager, error) {
	if sessionTTL <= 0 {
		sessionTTL = DefaultSessionTTL
	}

	manager := &AuthManager{
		credentialPath: CredentialPathForConfig(configFile),
		sessions:       make(map[string]sessionRecord),
		sessionTTL:     sessionTTL,
		nowFn:          time.Now,
		randReader:     rand.Reader,
	}

	if err := manager.loadCredentialsFromDisk(); err != nil {
		return nil, err
	}

	return manager, nil
}

// SessionTTL returns the configured session lifetime.
func (m *AuthManager) SessionTTL() time.Duration {
	return m.sessionTTL
}

// IsInitialized reports whether credentials exist.
func (m *AuthManager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.credentials != nil
}

// Username returns the configured username when initialized.
func (m *AuthManager) Username() (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.credentials == nil {
		return "", false
	}
	return m.credentials.Username, true
}

// Setup initializes single-user credentials.
func (m *AuthManager) Setup(username, password string) error {
	normalizedUsername := strings.TrimSpace(username)
	if normalizedUsername == "" {
		return fmt.Errorf("%w: username is required", ErrInvalidUsername)
	}
	if len(password) < minPasswordLength {
		return fmt.Errorf("%w: password must be at least %d characters", ErrWeakPassword, minPasswordLength)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.credentials != nil {
		return ErrAuthAlreadySet
	}

	params := argon2Params{
		Memory:  defaultArgon2Memory,
		Time:    defaultArgon2Time,
		Threads: defaultArgon2Threads,
		KeyLen:  defaultArgon2KeyLen,
	}

	salt, err := m.randomBytes(saltLength)
	if err != nil {
		return fmt.Errorf("failed to generate password salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, params.KeyLen)
	credentials := &storedCredentials{
		Username: normalizedUsername,
		Salt:     salt,
		Hash:     hash,
		Params:   params,
	}

	if err := m.writeCredentialsToDisk(credentials); err != nil {
		return err
	}

	m.credentials = credentials
	m.sessions = make(map[string]sessionRecord)
	m.resetFailedLoginStateLocked()

	return nil
}

// Login validates credentials and returns a new session ID.
func (m *AuthManager) Login(username, password string) (string, error) {
	normalizedUsername := strings.TrimSpace(username)

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.credentials == nil {
		return "", ErrAuthNotInitialized
	}

	now := m.nowFn()
	m.pruneExpiredSessionsLocked(now)
	if now.Before(m.loginBlockedUntil) {
		return "", ErrLoginRateLimited
	}

	if !m.verifyCredentialsLocked(normalizedUsername, password) {
		m.recordFailedLoginLocked(now)
		return "", ErrInvalidCredentials
	}
	m.resetFailedLoginStateLocked()

	sessionID, err := m.newSessionID()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	m.enforceSessionLimitLocked()

	m.sessions[sessionID] = sessionRecord{
		Username:  m.credentials.Username,
		ExpiresAt: now.Add(m.sessionTTL),
	}

	return sessionID, nil
}

// AuthenticateSession validates a session and returns its username.
func (m *AuthManager) AuthenticateSession(sessionID string) (string, error) {
	if strings.TrimSpace(sessionID) == "" {
		return "", ErrInvalidSession
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.credentials == nil {
		return "", ErrAuthNotInitialized
	}

	session, ok := m.sessions[sessionID]
	if !ok {
		return "", ErrInvalidSession
	}

	now := m.nowFn()
	if now.After(session.ExpiresAt) {
		delete(m.sessions, sessionID)
		return "", ErrInvalidSession
	}

	if session.Username != m.credentials.Username {
		delete(m.sessions, sessionID)
		return "", ErrInvalidSession
	}

	return session.Username, nil
}

// Logout invalidates a session.
func (m *AuthManager) Logout(sessionID string) {
	if strings.TrimSpace(sessionID) == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

func (m *AuthManager) recordFailedLoginLocked(now time.Time) {
	if m.failedLoginWindowStart.IsZero() || now.Sub(m.failedLoginWindowStart) > failedLoginWindow {
		m.failedLoginWindowStart = now
		m.failedLoginCount = 0
	}

	m.failedLoginCount++
	if m.failedLoginCount >= maxFailedLoginAttempts {
		m.loginBlockedUntil = now.Add(loginLockoutDuration)
		m.failedLoginCount = 0
		m.failedLoginWindowStart = time.Time{}
	}
}

func (m *AuthManager) resetFailedLoginStateLocked() {
	m.failedLoginCount = 0
	m.failedLoginWindowStart = time.Time{}
	m.loginBlockedUntil = time.Time{}
}

func (m *AuthManager) verifyCredentialsLocked(username, password string) bool {
	if m.credentials == nil {
		return false
	}

	passwordHash := argon2.IDKey(
		[]byte(password),
		m.credentials.Salt,
		m.credentials.Params.Time,
		m.credentials.Params.Memory,
		m.credentials.Params.Threads,
		m.credentials.Params.KeyLen,
	)

	usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(m.credentials.Username)) == 1
	passwordMatch := subtle.ConstantTimeCompare(passwordHash, m.credentials.Hash) == 1

	return usernameMatch && passwordMatch
}

func (m *AuthManager) pruneExpiredSessionsLocked(now time.Time) {
	for sessionID, session := range m.sessions {
		if now.After(session.ExpiresAt) {
			delete(m.sessions, sessionID)
		}
	}
}

func (m *AuthManager) enforceSessionLimitLocked() {
	for len(m.sessions) >= maxActiveSessions {
		var (
			oldestSessionID string
			oldestExpiry    time.Time
		)

		for sessionID, session := range m.sessions {
			if oldestSessionID == "" || session.ExpiresAt.Before(oldestExpiry) {
				oldestSessionID = sessionID
				oldestExpiry = session.ExpiresAt
			}
		}

		if oldestSessionID == "" {
			return
		}
		delete(m.sessions, oldestSessionID)
	}
}

func (m *AuthManager) newSessionID() (string, error) {
	raw, err := m.randomBytes(sessionIDBytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func (m *AuthManager) randomBytes(size int) ([]byte, error) {
	buf := make([]byte, size)
	if _, err := io.ReadFull(m.randReader, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (m *AuthManager) loadCredentialsFromDisk() error {
	if _, err := os.Lstat(m.credentialPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to stat auth credential file: %w", err)
	}

	if err := enforceCredentialFilePermissions(m.credentialPath); err != nil {
		return fmt.Errorf("failed to enforce auth credential permissions: %w", err)
	}

	data, err := os.ReadFile(m.credentialPath)
	if err != nil {
		return fmt.Errorf("failed to read auth credential file: %w", err)
	}

	var payload credentialFile
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to parse auth credential file: %w", err)
	}

	username := strings.TrimSpace(payload.Username)
	if username == "" {
		return fmt.Errorf("invalid auth credential file: username is required")
	}
	if payload.Memory == 0 || payload.Time == 0 || payload.Threads == 0 || payload.KeyLen == 0 {
		return fmt.Errorf("invalid auth credential file: argon2 parameters are required")
	}

	salt, err := base64.RawStdEncoding.DecodeString(payload.Salt)
	if err != nil || len(salt) == 0 {
		return fmt.Errorf("invalid auth credential file: invalid salt")
	}
	hash, err := base64.RawStdEncoding.DecodeString(payload.Hash)
	if err != nil || len(hash) == 0 {
		return fmt.Errorf("invalid auth credential file: invalid hash")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.credentials = &storedCredentials{
		Username: username,
		Salt:     salt,
		Hash:     hash,
		Params: argon2Params{
			Memory:  payload.Memory,
			Time:    payload.Time,
			Threads: payload.Threads,
			KeyLen:  payload.KeyLen,
		},
	}

	return nil
}

func (m *AuthManager) writeCredentialsToDisk(creds *storedCredentials) error {
	if creds == nil {
		return errors.New("credentials are required")
	}

	payload := credentialFile{
		Version:   1,
		Username:  creds.Username,
		Salt:      base64.RawStdEncoding.EncodeToString(creds.Salt),
		Hash:      base64.RawStdEncoding.EncodeToString(creds.Hash),
		Memory:    creds.Params.Memory,
		Time:      creds.Params.Time,
		Threads:   creds.Params.Threads,
		KeyLen:    creds.Params.KeyLen,
		CreatedAt: m.nowFn().UTC().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth credential file: %w", err)
	}

	dir := filepath.Dir(m.credentialPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create auth credential directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, credentialFilename+".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp auth credential file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write temp auth credential file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp auth credential file: %w", err)
	}

	if err := enforceCredentialFilePermissions(tmpPath); err != nil {
		return fmt.Errorf("failed to enforce temp auth credential permissions: %w", err)
	}

	if err := os.Rename(tmpPath, m.credentialPath); err != nil {
		return fmt.Errorf("failed to persist auth credential file: %w", err)
	}

	if err := enforceCredentialFilePermissions(m.credentialPath); err != nil {
		return fmt.Errorf("failed to enforce auth credential permissions: %w", err)
	}

	return nil
}
