package handler

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Context = event.Context[*Connection]

func (c *Connection) NotifyDisconnect(reason string) {
	c.handler.OnDisconnect(c, reason)
}

func (c *Connection) NotifyJoin() {
	c.handler.OnJoin(c)
}

func (c *Connection) replyNetworkStackLatency(pk *packet.NetworkStackLatency) {
	ctx := event.C(c)
	if c.handler.OnNetworkStackLatency(ctx, pk); ctx.Cancelled() {
		return
	}
	if !pk.NeedsResponse {
		return
	}
	c.requestNetworkStackLatency(pk)
}

func (c *Connection) replyLevelChunk(pk *packet.LevelChunk) {
	if pk.CacheEnabled {
		panic("ClientCache is Enabled.\n")
	}
	ctx := event.C(c)
	if c.handler.OnLevelChunk(ctx, pk); ctx.Cancelled() {
		return
	}
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().InsertLevelChunk(pk)
	})
}

func (c *Connection) replySubChunk(pk *packet.SubChunk) {
	if pk.CacheEnabled {
		panic("ClientCache is Enabled.\n")
	}
	ctx := event.C(c)
	if c.handler.OnSubChunk(ctx, pk); ctx.Cancelled() {
		return
	}
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().InsertSubChunk(pk)
	})
}

func (c *Connection) replyNetworkChunkPublisherUpdate(pk *packet.NetworkChunkPublisherUpdate) {
	ctx := event.C(c)
	if c.handler.OnNetworkChunkPublisherUpdate(ctx, pk); ctx.Cancelled() {
		return
	}
	posInMgl32 := utils.ProtocolPosToMgl32Vec3(pk.Position)
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().UpdateChunkRadius(int32(pk.Radius>>4))
		c.state.BlockMap().UpdateChunkCentre(posInMgl32)
	})
}

func (c *Connection) replyChunkRadiusUpdated(pk *packet.ChunkRadiusUpdated) {
	ctx := event.C(c)
	if c.handler.OnChunkRadiusUpdated(ctx, pk); ctx.Cancelled() {
		return
	}
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().UpdateChunkRadius(pk.ChunkRadius)
	})
}

func (c *Connection) replyUpdateAttributes(pk *packet.UpdateAttributes) {
	ctx := event.C(c)
	if c.handler.OnUpdateAttributes(ctx, pk); ctx.Cancelled() {
		return
	}
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		return
	}
	for _, attribute := range pk.Attributes {
		switch an := attribute.Name; an {
		case "minecraft:movement":
			c.state.Exec(func(q *game.Qx) {
				c.state.ReSimMoveAtTick(uint(pk.Tick), func(a *movements.AMovement) {
					a.BaseSpeed = attribute.Value
				})
			})
		}
	}
}

func (c *Connection) replySetActorMotion(pk *packet.SetActorMotion) {
	ctx := event.C(c)
	if c.handler.OnSetActorMotion(ctx, pk); ctx.Cancelled() {
		return
	}
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		return
	}
	c.state.Exec(func(q *game.Qx) {
		c.state.ReSimMoveAtTick(uint(pk.Tick), func(a *movements.AMovement) {
			a.Velocity = pk.Velocity
			a.Input.ServerSpeedAdd = mgl32.Vec3{}
		})
	})
}

func (c *Connection) replyUpdateBlock(pk *packet.UpdateBlock) {
	ctx := event.C(c)
	if c.handler.OnUpdateBlock(ctx, pk); ctx.Cancelled() {
		return
	}
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().SetBlock(pk.Position, uint8(pk.Layer), pk.NewBlockRuntimeID)
	})
}

func (c *Connection) replyMovePlayer(pk *packet.MovePlayer) {
	ctx := event.C(c)
	if c.handler.OnMovePlayer(ctx, pk); ctx.Cancelled() {
		return
	}
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		return
	}

	c.state.Exec(func(q *game.Qx) {
		c.state.ReSimMoveAtTick(uint(pk.Tick), func(a *movements.AMovement) {
			a.Position = pk.Position
			a.OnGround = pk.OnGround
			a.Input.Yaw = pk.Yaw
			a.Input.Pitch = pk.Pitch
		})
		c.state.BlockMap().UpdateChunkCentre(pk.Position)
		c.state.BlockMap().RefreshMapWithRenderDistance()
		if pk.Mode == packet.MoveModeTeleport{
			c.state.SetFlag(packet.InputFlagHandledTeleport)
		}
	})
}

func (c *Connection) replyCorrectPlayerMovePrediction(pk *packet.CorrectPlayerMovePrediction){
	ctx := event.C(c)
	if c.handler.OnCorrectPlayerMovePrediction(ctx, pk); ctx.Cancelled() {
		return
	}
	if pk.PredictionType == packet.PredictionTypePlayer{
		c.state.Exec(func(q *game.Qx) {
			c.state.ReSimMoveAtTick(uint(pk.Tick), func(a *movements.AMovement) {
				a.Position = pk.Position
				a.OnGround = pk.OnGround
				a.Input.Yaw = pk.Rotation[1]
				a.Input.Pitch = pk.Rotation[0]
			})
		})
	}
}