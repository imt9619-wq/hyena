package manager

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Manager struct {
	config       *Config
	accounts     []*Account
	accountConfMu *sync.Mutex
	sessionsMu   sync.Mutex
	sessions     map[uuid.UUID]*Session
	sessionQueue chan *Session
	closedNotify chan uuid.UUID
	closed       chan struct{}
	closeOnce    sync.Once
}

func (c Config) New(ctx context.Context) *Manager {
	m := &Manager{
		config:        &c,
		accountConfMu: &sync.Mutex{},
		closed:        make(chan struct{}),
		sessions:      make(map[uuid.UUID]*Session),
		sessionQueue:  make(chan *Session),
		closedNotify:  make(chan uuid.UUID, 64),
	}

	m.accountConfMu.Lock()
	accountConfigs := c.TokenStore.FetchAccounts()
	m.accountConfMu.Unlock()

	m.accounts = make([]*Account, 0, len(accountConfigs))
	for _, cfg := range accountConfigs {
		m.accounts = append(m.accounts, cfg.newAccount(m.sessionQueue, m.closedNotify, m.closed))
	}

	if ctx != nil {
		go m.closeOnContext(ctx)
	}
	go m.runSessionLoop()
	return m
}

func (m *Manager) Accounts() map[int]*Account {
	out := make(map[int]*Account, len(m.accounts))
	for i, acc := range m.accounts {
		out[i] = acc
	}
	return out
}

func (m *Manager) AccountsByTag() map[string]*Account {
	out := make(map[string]*Account, len(m.accounts))
	for _, acc := range m.accounts {
		out[acc.config.Tag] = acc
	}
	return out
}

func (m *Manager) closeOnContext(ctx context.Context) {
	<-ctx.Done()
	m.sessionsMu.Lock()
	empty := len(m.sessions) == 0
	m.sessionsMu.Unlock()
	if empty {
		m.Close()
	}
}

func (m *Manager) listenForSignals() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer signal.Stop(c)
		<-c
		m.Close()
	}()
}

func (m *Manager) Close() {
	m.closeOnce.Do(func() {
		m.sessionsMu.Lock()
		active := make([]*Session, 0, len(m.sessions))
		for _, s := range m.sessions {
			active = append(active, s)
		}
		m.sessionsMu.Unlock()

		for _, s := range active {
			_ = s.connection.WritePacket(&packet.Disconnect{})
			s.connection.NotifyDisconnect("Manager closed")
			s.markClosed()
		}
		close(m.closed)
	})
}

func (m *Manager) WaitTilClose() {
	m.listenForSignals()
	<-m.closed
}

func (m *Manager) runSessionLoop() {
	for {
		select {
		case session := <-m.sessionQueue:
			m.sessionsMu.Lock()
			m.sessions[session.id] = session
			m.sessionsMu.Unlock()

			go session.run()

			accountConfig := session.account.config
			go func(cfg *AccountConfig) {
				m.accountConfMu.Lock()
				m.config.TokenStore.SaveAccount(*cfg)
				m.accountConfMu.Unlock()
			}(accountConfig)

		case sessionID := <-m.closedNotify:
			m.sessionsMu.Lock()
			delete(m.sessions, sessionID)
			empty := len(m.sessions) == 0
			m.sessionsMu.Unlock()
			if empty {
				m.Close()
			}

		case <-m.closed:
			return
		}
	}
}
