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
	conf          *Config
	clients       []*Client
	clientConfMu  *sync.Mutex
	connsMu       sync.Mutex
	conns         map[uuid.UUID]*ClientConn
	incomingConns chan *ClientConn
	closingConns chan uuid.UUID
	closed        chan struct{}
	closeOnce     sync.Once
}

func (c Config) New(ctx context.Context) *Manager {
	m := &Manager{
		conf:          &c,
		clientConfMu:  &sync.Mutex{},
		closed:        make(chan struct{}),
		conns:         make(map[uuid.UUID]*ClientConn),
		incomingConns: make(chan *ClientConn),
		closingConns: make(chan uuid.UUID, 64),
	}

	m.clientConfMu.Lock()
	clientConfs := c.CachedClients.FetchClients()
	m.clientConfMu.Unlock()

	m.clients = make([]*Client, 0, len(clientConfs))
	for _, clientConf := range clientConfs {
		m.clients = append(m.clients, clientConf.new(m.incomingConns, m.closingConns, m.closed))
	}

	if ctx != nil {
		go m.closeWithContext(ctx)
	}
	go m.startTakingConn()
	return m
}

func (mgr *Manager) Clients() map[int]*Client {
	indexToClient := make(map[int]*Client, len(mgr.clients))
	for index, client := range mgr.clients {
		indexToClient[index] = client
	}
	return indexToClient
}

func (mgr *Manager) closeWithContext(ctx context.Context) {
	<-ctx.Done()
	mgr.connsMu.Lock()
	empty := len(mgr.conns) == 0
	mgr.connsMu.Unlock()
	if empty {
		mgr.Close()
	}
}

func (mgr *Manager) closeManagerOnEnd() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer signal.Stop(c)
		<-c
		mgr.Close()
	}()
}

func (mgr *Manager) Close() {
	mgr.closeOnce.Do(func() {
		mgr.connsMu.Lock()
		active := make([]*ClientConn, 0, len(mgr.conns))
		for _, cc := range mgr.conns {
			active = append(active, cc)
		}
		mgr.connsMu.Unlock()

		for _, cc := range active {
			_ = cc.connBuf.WritePacket(&packet.Disconnect{})
			cc.markClosed()
		}
		close(mgr.closed)
	})
}

func (mgr *Manager) WaitTilClose() {
	mgr.closeManagerOnEnd()
	<-mgr.closed
}


func (mgr *Manager) startTakingConn() {
	for {
		select{
		case newcc := <- mgr.incomingConns:
			go newcc.handleConn()
			mgr.connsMu.Lock()
			mgr.conns[newcc.id] = newcc
			mgr.connsMu.Unlock()
			clientConf := newcc.client.conf

			go func(conf *ClientConfig){
				mgr.clientConfMu.Lock()
				mgr.conf.CachedClients.SaveClients(*conf)
				mgr.clientConfMu.Unlock()
			}(clientConf)

		case closeccId := <- mgr.closingConns:
			mgr.connsMu.Lock()
			delete(mgr.conns, closeccId)
			empty := len(mgr.conns) == 0
			mgr.connsMu.Unlock()
			if empty {
				mgr.Close()
			}
			
		case <-mgr.closed: 
			return
		}
	}
}
