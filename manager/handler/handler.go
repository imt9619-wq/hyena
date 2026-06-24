package handler

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Handler receives connection lifecycle events for a server session.
type Handler interface {
	OnDisconnect(*Connection, string)
	OnJoin(*Connection)
	OnNetworkStackLatency(*Context, *packet.NetworkStackLatency)
	OnMoveActorAbsolute(*Context, *packet.MoveActorAbsolute)
	OnLevelChunk(*Context, *packet.LevelChunk)
	OnSubChunk(*Context, *packet.SubChunk)
	OnNetworkChunkPublisherUpdate(*Context, *packet.NetworkChunkPublisherUpdate)
	OnChunkRadiusUpdated(*Context, *packet.ChunkRadiusUpdated)
	OnUpdateAttributes(*Context, *packet.UpdateAttributes)
	OnSetActorMotion(*Context, *packet.SetActorMotion)
	OnUpdateBlock(*Context, *packet.UpdateBlock)
	OnMovePlayer(*Context, *packet.MovePlayer)
}

type DefaultHandler struct{}

var _ Handler = DefaultHandler{}

func (h DefaultHandler) OnDisconnect(c *Connection, reason string) {
	fmt.Printf("%s disconnected: %s\n", c.IdentityData().DisplayName, reason)
}

func (h DefaultHandler) OnJoin(c *Connection) {
	fmt.Printf("%s has joined the server: %s\n", c.IdentityData().DisplayName, c.RemoteAddr())
	c.StartRunning()
	c.StartJumping()
}

func (h DefaultHandler) OnNetworkStackLatency(*Context, *packet.NetworkStackLatency){}
func (h DefaultHandler) OnMoveActorAbsolute(*Context, *packet.MoveActorAbsolute){}
func (h DefaultHandler) OnLevelChunk(*Context, *packet.LevelChunk){}
func (h DefaultHandler) OnSubChunk(*Context, *packet.SubChunk){}
func (h DefaultHandler) OnNetworkChunkPublisherUpdate(*Context, *packet.NetworkChunkPublisherUpdate){}
func (h DefaultHandler) OnChunkRadiusUpdated(*Context, *packet.ChunkRadiusUpdated){}
func (h DefaultHandler) OnUpdateAttributes(*Context, *packet.UpdateAttributes){}
func (h DefaultHandler) OnSetActorMotion(*Context, *packet.SetActorMotion){}
func (h DefaultHandler) OnUpdateBlock(*Context, *packet.UpdateBlock){}
func (h DefaultHandler) OnMovePlayer(*Context, *packet.MovePlayer){}