package manager

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/imt9619-wq/hyena/manager/handler"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type ClientConn struct {
	connBuf *handler.ConnBuf
	client    *Client
	id        uuid.UUID
	closeOnce sync.Once
}

func (cc *ClientConn) markClosed() {
	cc.closeOnce.Do(func() {
		cc.connBuf.Close()
		
		select {
		case cc.client.closeConnChan <- cc.id:
		case <-cc.client.managerClosed:
		default:
		}

	})
}

func (cc *ClientConn) handleConn() {
	defer cc.markClosed()
	serverConn := cc.connBuf
	serverConn.BhJoin()

	for {
		pk, err := serverConn.ReadPacket()
		if err != nil {
			var disc minecraft.DisconnectError
			if errors.As(err, &disc) {
				serverConn.BhDisconnect(disc.Error())
			}
			return
		}

		switch pk := pk.(type) {
		case *packet.StartGame:
			serverConn.BhStartGame(pk)
		case *packet.NetworkStackLatency:
			serverConn.BhNetworkStackLatency(pk)
		case *packet.MoveActorAbsolute:
			serverConn.BhMoveActorAbsolute(pk)
		default:
		}
	}
}

func (cc *ClientConn) Handle(h handler.ConnHandler) {
	cc.connBuf.Handle(h)
}

