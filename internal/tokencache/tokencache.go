package tokencache

import (
	"encoding/json"
	"os"
	"path/filepath"
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

var openKeyring = defaultOpenKeyring

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

func Save(token string, expiresAt time.Time, userID string) error {
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
		Key:   tokenKey,
		Label: serviceName + " token",
		Data:  data,
	}); err != nil {
		log.Debug("keyring set failed", "error", err)
		return err
	}
	log.Debug("keyring saved token")
	return nil
}

func Load() (*CachedToken, error) {
	ring, err := openKeyring()
	if err != nil {
		log.Debug("keyring open failed (load)", "error", err)
		return nil, err
	}
	item, err := ring.Get(tokenKey)
	if err != nil {
		log.Debug("keyring get failed", "error", err)
		return nil, err
	}
	var cached CachedToken
	if err := json.Unmarshal(item.Data, &cached); err != nil {
		return nil, err
	}
	if time.Now().After(cached.ExpiresAt) {
		return nil, keyring.ErrKeyNotFound
	}
	return &cached, nil
}

func Clear() error {
	ring, err := openKeyring()
	if err != nil {
		return err
	}
	return ring.Remove(tokenKey)
}
