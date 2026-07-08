package handler

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/imt9619-wq/hyena/manager/handler/form"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
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
	if !pk.NeedsResponse {
		return
	}
	c.requestNetworkStackLatency(pk)
}

func (c *Connection) replyLevelChunk(pk *packet.LevelChunk) {
	if pk.CacheEnabled {
		panic("ClientCache is Enabled.\n")
	}
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().InsertLevelChunk(pk)
	})
}

func (c *Connection) replySubChunk(pk *packet.SubChunk) {
	if pk.CacheEnabled {
		panic("ClientCache is Enabled.\n")
	}
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().InsertSubChunk(pk)
	})
}

func (c *Connection) replyNetworkChunkPublisherUpdate(pk *packet.NetworkChunkPublisherUpdate) {
	posInMgl32 := utils.ProtocolPosToMgl32Vec3(pk.Position)
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().UpdateChunkRadius(int32(pk.Radius>>4))
		c.state.BlockMap().UpdateChunkCentre(posInMgl32)
	})
}

func (c *Connection) replyChunkRadiusUpdated(pk *packet.ChunkRadiusUpdated) {
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().UpdateChunkRadius(pk.ChunkRadius)
	})
}

func (c *Connection) replyUpdateAttributes(pk *packet.UpdateAttributes) {
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		return
	}
	for _, attribute := range pk.Attributes {
		switch an := attribute.Name; an {
		case "minecraft:movement":
			c.state.Exec(func(q *game.Qx) {
				c.state.ReSimMoveAtTick(uint(pk.Tick), func(im *movements.InMovement) {
					im.BaseSpeed = attribute.Value
				})
			})
		}
	}
}

func (c *Connection) replySetActorMotion(pk *packet.SetActorMotion) {
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		return
	}
	c.state.Exec(func(q *game.Qx) {
		c.state.ReSimMoveAtTick(uint(pk.Tick), func(im *movements.InMovement) {
			im.Velocity = pk.Velocity
			im.Input.ServerSpeedAdd = mgl32.Vec3{}
		})
	})
}

func (c *Connection) replyUpdateBlock(pk *packet.UpdateBlock) {
	c.state.Exec(func(q *game.Qx) {
		c.state.BlockMap().SetBlock(pk.Position, uint8(pk.Layer), pk.NewBlockRuntimeID)
	})
}

func (c *Connection) replyMovePlayer(pk *packet.MovePlayer) {
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		return
	}

	c.state.Exec(func(q *game.Qx) {
		c.state.ReSimMoveAtTick(uint(pk.Tick), func(im *movements.InMovement) {
			im.Position = pk.Position
			im.OnGround = pk.OnGround
			im.Input.Yaw = pk.Yaw
			im.Input.Pitch = pk.Pitch
		})
		c.state.BlockMap().UpdateChunkCentre(pk.Position)
		c.state.BlockMap().RefreshMapWithRenderDistance()
		if pk.Mode == packet.MoveModeTeleport{
			c.state.SetFlag(packet.InputFlagHandledTeleport)
		}
	})
}

func (c *Connection) replyCorrectPlayerMovePrediction(pk *packet.CorrectPlayerMovePrediction){
	if pk.PredictionType == packet.PredictionTypePlayer{
		c.state.Exec(func(q *game.Qx) {
			c.state.ReSimMoveAtTick(uint(pk.Tick), func(im *movements.InMovement) {
				im.Position = pk.Position
				im.OnGround = pk.OnGround
				im.Input.Yaw = pk.Rotation[1]
				im.Input.Pitch = pk.Rotation[0]
			})
		})
	}
}

func (c *Connection) replyInventoryContent(pk *packet.InventoryContent){
	c.state.Exec(func(q *game.Qx) {
		c.state.Inventory().SyncInventoryContent(pk)
	})
}

func (c *Connection) replyMobEquipment(pk *packet.MobEquipment){
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		return
	}
	c.state.Exec(func(q *game.Qx) {
		c.state.Inventory().Equip(pk)
	})
}

func (c *Connection) replyModalFormRequest(pk *packet.ModalFormRequest){
	ctx := event.C(c)
	f, ok := form.UnmarshalForm(pk.FormID, pk.FormData)
	if !ok{
		return
	}
	cancelForm := func ()  {
		c.WritePacket(&packet.ModalFormResponse{
			FormID: pk.FormID,
			CancelReason: protocol.Option(uint8(packet.ModalFormCancelReasonUserClosed)),
		})
	}
	if c.handler.OnForm(ctx, f); ctx.Cancelled(){
		cancelForm()
		return
	}
	data := f.ResponseJson()
	if data == nil{
		cancelForm()
		return
	}
	c.WritePacket(&packet.ModalFormResponse{
		FormID: pk.FormID,
		ResponseData: protocol.Option(data),
	})
}