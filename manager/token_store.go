package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
)

// TokenStore loads and persists account OAuth tokens.
type TokenStore interface {
	FetchAccounts() []AccountConfig
	SaveAccount(AccountConfig)
}

// FileTokenStore reads and writes token JSON files from a directory.
type FileTokenStore struct {
	Dir string
}

func (s *FileTokenStore) FetchAccounts() []AccountConfig {
	accounts := make([]AccountConfig, 0, 10)

	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		log.Fatal(err)
	}

	entries, _ := os.ReadDir(s.Dir)
	for _, entry := range entries {
		path := filepath.Join(s.Dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil || !json.Valid(data) {
			continue
		}

		var tok oauth2.Token
		if err := json.Unmarshal(data, &tok); err != nil {
			fmt.Printf("error parsing token %s: %v\n", entry.Name(), err)
			continue
		}
		tag := strings.TrimSuffix(entry.Name(), ".json")
		accounts = append(accounts, AccountConfig{Tag: tag, Token: &tok})
	}

	return accounts
}

func (s *FileTokenStore) SaveAccount(cfg AccountConfig) {
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		log.Fatal(err)
	}
	if cfg.Token == nil {
		return
	}

	path := filepath.Join(s.Dir, cfg.Tag+".json")
	_ = os.MkdirAll(filepath.Dir(path), 0o700)

	b, err := json.Marshal(cfg.Token)
	if err != nil {
		return
	}
	_ = os.WriteFile(path, b, 0o600)
	fmt.Printf("%s token saved\n", cfg.Tag)
}
