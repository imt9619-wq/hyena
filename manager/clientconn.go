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
		_ = cc.connBuf.Close()
		
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
	h := serverConn.H
	if h == nil {
		h = handler.DefaultHandler{}
		serverConn.H = h
	}
	for {
		pk, err := serverConn.ReadPacket()
		if err != nil {
			var disc minecraft.DisconnectError
			if errors.As(err, &disc) {
				h.HandleDisconnect(serverConn, disc.Error())
			}
			return
		}

		switch pk := pk.(type) {
		case *packet.NetworkStackLatency:
			serverConn.BhNSL(pk)
		default:
		}
	}
}

func (cc *ClientConn) Handle(h handler.ConnHandler) {
	cc.connBuf.Handle(h)
}

