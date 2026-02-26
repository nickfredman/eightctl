package tokencache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"github.com/charmbracelet/log"
)

const (
	serviceName = "eightctl"
	tokenKey    = "oauth-token"
)

type CachedToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id,omitempty"`
}

// Identity describes the authentication context a token belongs to.
// Tokens are namespaced by base URL, client ID, and email so switching
// between accounts or environments doesn't reuse the wrong credentials.
type Identity struct {
	BaseURL  string
	ClientID string
	Email    string
}

var openKeyring = defaultOpenKeyring

// SetOpenKeyringForTest swaps the keyring opener; it returns a restore func.
// Not safe for concurrent tests; intended for isolated test scenarios.
func SetOpenKeyringForTest(fn func() (keyring.Keyring, error)) (restore func()) {
	prev := openKeyring
	openKeyring = fn
	return func() { openKeyring = prev }
}

func defaultOpenKeyring() (keyring.Keyring, error) {
	home, _ := os.UserHomeDir()
	return keyring.Open(keyring.Config{
		ServiceName: serviceName,
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.SecretServiceBackend,
			keyring.WinCredBackend,
			keyring.FileBackend,
		},
		FileDir:          filepath.Join(home, ".config", "eightctl", "keyring"),
		FilePasswordFunc: filePassword,
	})
}

func filePassword(_ string) (string, error) {
	return serviceName + "-fallback", nil
}

func Save(id Identity, token string, expiresAt time.Time, userID string) error {
	ring, err := openKeyring()
	if err != nil {
		log.Debug("keyring open failed (save)", "error", err)
		return err
	}
	data, err := json.Marshal(CachedToken{
		Token:     token,
		ExpiresAt: expiresAt,
		UserID:    userID,
	})
	if err != nil {
		return err
	}
	if err := ring.Set(keyring.Item{
		Key:   cacheKey(id),
		Label: serviceName + " token",
		Data:  data,
	}); err != nil {
		log.Debug("keyring set failed", "error", err)
		return err
	}
	log.Debug("keyring saved token")
	return nil
}

func Load(id Identity, expectedUserID string) (*CachedToken, error) {
	ring, err := openKeyring()
	if err != nil {
		log.Debug("keyring open failed (load)", "error", err)
		return nil, err
	}
	key := cacheKey(id)
	item, err := ring.Get(key)
	if err == keyring.ErrKeyNotFound {
		// Backward/compatibility lookups when the exact identity key is missing.
		if id.Email == "" {
			// No email specified: attempt to find a single matching token for this base/client.
			if alt, findErr := findSingleForClient(ring, id); findErr == nil {
				key = alt
				item, err = ring.Get(key)
			} else {
				log.Debug("keyring wildcard lookup failed", "error", findErr)
			}
		} else {
			// Try legacy key shapes first (some older builds used no client-id in key).
			if legacyKey, ok := tryLegacyKeys(ring, id); ok {
				key = legacyKey
				item, err = ring.Get(key)
			} else if alt, findErr := findSingleForBaseAndEmail(ring, id); findErr == nil {
				// Client id may have changed across releases; recover any token for
				// same base URL + email to avoid forced re-login/rate-limit loops.
				key = alt
				item, err = ring.Get(key)
			} else {
				log.Debug("keyring compat lookup failed", "error", findErr)
			}
		}
	}
	if err != nil {
		log.Debug("keyring get failed", "error", err)
		return nil, err
	}
	var cached CachedToken
	if err := json.Unmarshal(item.Data, &cached); err != nil {
		return nil, err
	}
	if time.Now().After(cached.ExpiresAt) {
		_ = ring.Remove(key)
		return nil, keyring.ErrKeyNotFound
	}
	if expectedUserID != "" && cached.UserID != "" && cached.UserID != expectedUserID {
		return nil, keyring.ErrKeyNotFound
	}
	return &cached, nil
}

func Clear(id Identity) error {
	ring, err := openKeyring()
	if err != nil {
		return err
	}
	key := cacheKey(id)
	if err := ring.Remove(key); err != nil {
		if err == keyring.ErrKeyNotFound || os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

func cacheKey(id Identity) string {
	base := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(id.BaseURL)), "/")
	email := strings.ToLower(strings.TrimSpace(id.Email))
	return tokenKey + ":" + base + "|" + id.ClientID + "|" + email
}

func tryLegacyKeys(ring keyring.Keyring, id Identity) (string, bool) {
	base := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(id.BaseURL)), "/")
	email := strings.ToLower(strings.TrimSpace(id.Email))
	if email == "" {
		return "", false
	}
	candidates := []string{
		tokenKey + ":" + base + "||" + email,
		tokenKey + ":" + base + "|" + email,
		tokenKey + ":" + email,
	}
	for _, k := range candidates {
		if _, err := ring.Get(k); err == nil {
			return k, true
		}
	}
	return "", false
}

// findSingleForClient finds a single cached key for the given base/client when email is unknown.
// Returns ErrKeyNotFound if none or multiple exist.
func findSingleForClient(ring keyring.Keyring, id Identity) (string, error) {
	keys, err := ring.Keys()
	if err != nil {
		return "", err
	}
	prefix := tokenKey + ":" + strings.TrimSuffix(strings.ToLower(strings.TrimSpace(id.BaseURL)), "/") + "|" + id.ClientID + "|"
	matches := []string{}
	for _, k := range keys {
		if strings.HasPrefix(k, prefix) {
			matches = append(matches, k)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	return "", keyring.ErrKeyNotFound
}

// findSingleForBaseAndEmail finds a token key for the same base URL + email,
// regardless of client id (for compatibility across releases).
func findSingleForBaseAndEmail(ring keyring.Keyring, id Identity) (string, error) {
	keys, err := ring.Keys()
	if err != nil {
		return "", err
	}
	base := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(id.BaseURL)), "/")
	email := strings.ToLower(strings.TrimSpace(id.Email))
	suffix := "|" + email
	prefix := tokenKey + ":" + base + "|"
	matches := []string{}
	for _, k := range keys {
		if strings.HasPrefix(k, prefix) && strings.HasSuffix(k, suffix) {
			matches = append(matches, k)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	return "", keyring.ErrKeyNotFound
}
