package manager

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/imt9619-wq/hyena/manager/handler"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
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
			}
			return
		}

		switch pk := pk.(type) {
		case *packet.NetworkStackLatency:
			conn.ReplyNetworkStackLatency(pk)
		case *packet.MoveActorAbsolute:
			conn.ReplyMoveActorAbsolute(pk)
		case *packet.LevelChunk:
			conn.ReplyLevelChunk(pk)
		case *packet.NetworkChunkPublisherUpdate:
			conn.ReplyNetworkChunkPublisherUpdate(pk)
		case *packet.ChunkRadiusUpdated:
			conn.ReplyChunkRadiusUpdated(pk)
		case *packet.UpdateAttributes:
			conn.ReplyUpdateAttributes(pk)
		case *packet.SetActorMotion:
			conn.ReplySetActorMotion(pk)
		case *packet.UpdateBlock:
			conn.ReplyUpdateBlock(pk)
		default:
		}
	}
}

func (s *Session) SetHandler(h handler.Handler) {
	s.connection.SetHandler(h)
}
