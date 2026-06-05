package handler

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (c *Connection) NotifyDisconnect(reason string) {
	c.handler.OnDisconnect(c, reason)
}

func (c *Connection) NotifyJoin() {
	c.handler.OnJoin(c)
}

func (c *Connection) ReplyNetworkStackLatency(pk *packet.NetworkStackLatency) {
	if !pk.NeedsResponse {
		return
	}
	c.WritePacket(&packet.NetworkStackLatency{
		Timestamp:     pk.Timestamp * 1000000,
		NeedsResponse: pk.NeedsResponse,
	})
}

func (c *Connection) SyncActorPosition(pk *packet.MoveActorAbsolute) {
	if c.state.entityRuntimeID != pk.EntityRuntimeID {
		return
	}
	yaw, pitch := rotationToPitchAndYaw(pk.Rotation)
	ps := c.state.player
	ps.Lock()
	defer ps.Unlock()

	ps.position = pk.Position
	ps.velocity = mgl32.Vec3{}
	ps.pitch = pitch
	ps.yaw = yaw
}

func (c *Connection) MapBlocks(pk *packet.LevelChunk){
	
}
