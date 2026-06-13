package handler

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game"
	"github.com/imt9619-wq/hyena/game/blockmap"
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
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		return
	}
	yaw, pitch := rotationToPitchAndYaw(pk.Rotation)
	ps := c.state.Player()

	c.state.Exec(func(q *game.Qx) {
		ps.Position = pk.Position
		ps.Velocity = mgl32.Vec3{}
		ps.Pitch = pitch
		ps.Yaw = yaw
	})
}

func (c *Connection) ReplyLevelChunk(pk *packet.LevelChunk) {
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().InsertLevelChunk(pk)
	})
}

func (c *Connection) ReplyNetworkChunkPublisherUpdate(pk *packet.NetworkChunkPublisherUpdate) {
	posInMgl32 := blockmap.ProtocolPosToMgl32Vec3(pk.Position)
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().UpdateChunkRadius(int32(pk.Radius))
		c.state.BlockMap().UpdateChunkCentre(posInMgl32)
	})
}

func (c *Connection) ReplyChunkRadiusUpdated(pk *packet.ChunkRadiusUpdated) {
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().UpdateChunkRadius(pk.ChunkRadius)
	})
}

func (c *Connection) ReplyUpdateAttributes(pk *packet.UpdateAttributes) {
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		return
	}
	for _, attribute := range pk.Attributes {
		switch an := attribute.Name; an {
		case "minecraft:movement":
			c.state.Exec(func(q *game.Qx) {
				c.state.Player().SetSpeedTo(attribute.Value)
			})
		}
	}
}

func (c *Connection) ReplySetActorMotion(pk *packet.SetActorMotion) {
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		return
	}
	c.state.Exec(func(q *game.Qx) {
		c.state.Player().Velocity = pk.Velocity
	})
}

func (c *Connection) ReplyUpdateBlock(pk *packet.UpdateBlock) {
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().SetBlock(pk.Position, uint8(pk.Layer), pk.NewBlockRuntimeID)
	})
}
