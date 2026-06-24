package manager

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/imt9619-wq/hyena/manager/handler"
	"github.com/sandertv/gophertunnel/minecraft"
)

// Session is an active connection to a Minecraft server for one account.
type Session struct {
	connection *handler.Connection
	account    *Account
	id         uuid.UUID
	closeOnce  sync.Once
}

func (s *Session) markClosed() {
	s.closeOnce.Do(func() {
		s.connection.Close()

		select {
		case s.account.closedNotify <- s.id:
		case <-s.account.managerClosed:
		default:
		}
	})
}

func (s *Session) run() {
	defer s.markClosed()

	conn := s.connection
	conn.NotifyJoin()

	for {
		pk, err := conn.ReadPacket()
		if err != nil {
			var disc minecraft.DisconnectError
			if errors.As(err, &disc) {
				conn.NotifyDisconnect(disc.Error())
			}else{
				conn.NotifyDisconnect(fmt.Sprintf("Error occured: %v\n", err))
			}
			return
		}
		conn.HandlePacket(pk)
	}
}

func (s *Session) SetHandler(h handler.Handler) {
	s.connection.SetHandler(h)
}
