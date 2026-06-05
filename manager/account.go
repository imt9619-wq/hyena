package manager

import (
	"github.com/google/uuid"
	"github.com/imt9619-wq/hyena/manager/handler"
	"github.com/sandertv/gophertunnel/minecraft"
	"golang.org/x/oauth2"
)

// Account represents a Minecraft account backed by a stored OAuth token.
type Account struct {
	config        *AccountConfig
	id            uuid.UUID
	sessionQueue  chan *Session
	closedNotify  chan uuid.UUID
	managerClosed <-chan struct{}
}

// AccountConfig holds the persisted identity for one account.
type AccountConfig struct {
	Tag   string // filename tag, e.g. "ms_token_cache" from ms_token_cache.json
	Token *oauth2.Token
}

func (cfg AccountConfig) newAccount(sessionQueue chan *Session, closedNotify chan uuid.UUID, managerClosed <-chan struct{}) *Account {
	return &Account{
		config:        &cfg,
		id:            uuid.New(),
		sessionQueue:  sessionQueue,
		closedNotify:  closedNotify,
		managerClosed: managerClosed,
	}
}

func (a *Account) newSession(conn *minecraft.Conn, h handler.Handler) *Session {
	return &Session{
		connection: handler.NewConnection(conn, h),
		account:    a,
		id:         uuid.New(),
	}
}
