package handler

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/manager/handler/blockmap"
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

func (c *Connection) ReplyMoveActorAbsolute(pk *packet.MoveActorAbsolute) {
	if c.state.entityRuntimeID != pk.EntityRuntimeID {
		return
	}
	yaw, pitch := rotationToPitchAndYaw(pk.Rotation)
	ps := c.state.player

	c.state.Exec(func(q *Qx) {
		ps.position = pk.Position
		ps.velocity = mgl32.Vec3{}
		ps.pitch = pitch
		ps.yaw = yaw
	})
}

func (c *Connection) ReplyLevelChunk(pk *packet.LevelChunk) {
	c.state.Exec(func(q *Qx) {
		c.state.blockMap.InsertLevelChunk(pk)
	})
}

func (c *Connection) ReplyNetworkChunkPublisherUpdate(pk *packet.NetworkChunkPublisherUpdate) {
	posInMgl32 := blockmap.ProtocolPosToMgl32Vec3(pk.Position)
	c.state.Exec(func(q *Qx) {
		c.state.blockMap.UpdateChunkRadius(int32(pk.Radius))
		c.state.blockMap.UpdateChunkCentre(posInMgl32)
	})
}

func (c *Connection) ReplyChunkRadiusUpdated(pk *packet.ChunkRadiusUpdated) {
	c.state.Exec(func(q *Qx) {
		c.state.blockMap.UpdateChunkRadius(pk.ChunkRadius)
	})
}

func (c *Connection) ReplyUpdateAttributes(pk *packet.UpdateAttributes) {
	if c.state.entityRuntimeID != pk.EntityRuntimeID {
		return
	}
	for _, attribute := range pk.Attributes {
		switch an := attribute.Name; an {
		case "minecraft:movement":
			c.state.Exec(func(q *Qx) {
				c.state.player.setSpeedTo(attribute.Value)
			})
		}
	}
}

func (c *Connection) ReplySetActorMotion(pk *packet.SetActorMotion) {
	if c.state.entityRuntimeID != pk.EntityRuntimeID {
		return
	}
	c.state.Exec(func(q *Qx) {
		c.state.player.velocity = pk.Velocity
	})
}

func (c *Connection) ReplyUpdateBlock(pk *packet.UpdateBlock) {
	c.state.Exec(func(q *Qx) {
		c.state.blockMap.SetBlock(pk.Position, uint8(pk.Layer), pk.NewBlockRuntimeID)
	})
}
