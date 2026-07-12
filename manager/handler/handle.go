package handler

import (
	"time"

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

func (c *Connection) handleNetworkStackLatency(pk *packet.NetworkStackLatency) {
	if !pk.NeedsResponse {
		return
	}
	c.requestNetworkStackLatency(pk)
}

func (c *Connection) handleLevelChunk(pk *packet.LevelChunk) {
	if pk.CacheEnabled {
		panic("ClientCache is Enabled.\n")
	}
	c.state.Exec(func(q *game.Qx) {
		q.InsertLevelChunk(pk)
	})
}

func (c *Connection) handleSubChunk(pk *packet.SubChunk) {
	if pk.CacheEnabled {
		panic("ClientCache is Enabled.\n")
	}
	c.state.Exec(func(q *game.Qx) {
		q.InsertSubChunk(pk)
	})
}

func (c *Connection) handleNetworkChunkPublisherUpdate(pk *packet.NetworkChunkPublisherUpdate) {
	posInMgl32 := utils.ProtocolPosToMgl32Vec3(pk.Position)
	c.state.Exec(func(q *game.Qx) {
		q.UpdateChunkRadius(int32(pk.Radius>>4))
		q.UpdateChunkCentre(posInMgl32)
	})
}

func (c *Connection) handleChunkRadiusUpdated(pk *packet.ChunkRadiusUpdated) {
	c.state.Exec(func(q *game.Qx) {
		q.UpdateChunkRadius(pk.ChunkRadius)
	})
}

func (c *Connection) handleUpdateAttributes(pk *packet.UpdateAttributes) {
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		return
	}
	for _, attribute := range pk.Attributes {
		switch an := attribute.Name; an {
		case "minecraft:movement":
			c.state.Exec(func(q *game.Qx) {
				q.ResimMove(uint(pk.Tick), func(im *movements.InMovement) {
					im.BaseSpeed = attribute.Value
				})
			})
		}
	}
}

func (c *Connection) handleSetActorMotion(pk *packet.SetActorMotion) {
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		return
	}
	c.state.Exec(func(q *game.Qx) {
		q.ResimMove(uint(pk.Tick), func(im *movements.InMovement) {
			im.Velocity = pk.Velocity
			im.Input.ServerSpeedAdd = mgl32.Vec3{}
		})
	})
}

func (c *Connection) handleUpdateBlock(pk *packet.UpdateBlock) {
	c.state.Exec(func(q *game.Qx) {
		q.SetBlock(pk.Position, uint8(pk.Layer), pk.NewBlockRuntimeID)
	})
}

func (c *Connection) handleUpdateSubChunkBlocks(pk *packet.UpdateSubChunkBlocks){
	c.state.Exec(func(q *game.Qx) {
		for _, bEntry := range pk.Blocks{
			q.SetBlock(utils.ProtocolBlockPosAdd(pk.Position, bEntry.BlockPos), 0, bEntry.BlockRuntimeID)
		}
		for _, bEntry := range pk.Extra{
			q.SetBlock(utils.ProtocolBlockPosAdd(pk.Position, bEntry.BlockPos), 1, bEntry.BlockRuntimeID)
		}
	})
}

func (c *Connection) handleMovePlayer(pk *packet.MovePlayer) {
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		c.entInWorld.movePlayer(pk)
		return
	}

	c.state.Exec(func(q *game.Qx) {
		q.ResimMove(uint(pk.Tick), func(im *movements.InMovement) {
			im.Position = pk.Position
			im.OnGround = pk.OnGround
			im.Input.Yaw = pk.Yaw
			im.Input.Pitch = pk.Pitch
		})
		q.UpdateChunkCentre(pk.Position)
		if pk.Mode == packet.MoveModeTeleport{
			q.SetInputFlag(packet.InputFlagHandledTeleport)
		}
	})
}

func (c *Connection) handleCorrectPlayerMovePrediction(pk *packet.CorrectPlayerMovePrediction){
	if pk.PredictionType == packet.PredictionTypePlayer{
		c.state.Exec(func(q *game.Qx) {
			q.ResimMove(uint(pk.Tick), func(im *movements.InMovement) {
				im.Position = pk.Position
				im.OnGround = pk.OnGround
				im.Input.Yaw = pk.Rotation[1]
				im.Input.Pitch = pk.Rotation[0]
			})
		})
	}
}

func (c *Connection) handleInventoryContent(pk *packet.InventoryContent){
	c.state.Exec(func(q *game.Qx) {
		q.SyncInventoryContent(pk)
	})
}

func (c *Connection) handleMobEquipment(pk *packet.MobEquipment){
	if c.state.EntityRunTimeId() != pk.EntityRuntimeID {
		return
	}
	c.state.Exec(func(q *game.Qx) {
		q.Equip(pk)
	})
}

func (c *Connection) handleModalFormRequest(pk *packet.ModalFormRequest){
	ctx := event.C(c)
	f, ok := form.UnmarshalForm(pk.FormID, pk.FormData)
	if !ok{
		return
	}
	cancelForm := func () *packet.ModalFormResponse{
		return &packet.ModalFormResponse{
			FormID: pk.FormID,
			CancelReason: protocol.Option(uint8(packet.ModalFormCancelReasonUserClosed)),
		}
	}
	delayWrite := func (pk packet.Packet){
		go func ()  {
			time.Sleep(500*time.Millisecond)
			c.onUi.Store(false)
			c.WritePacket(pk)
		}()
	}
	c.onUi.Store(true)
	if c.handler.OnForm(ctx, f); ctx.Cancelled(){
		delayWrite(cancelForm())
		return
	}
	data := f.ResponseJson()
	if data == nil{
		delayWrite(cancelForm())
		return
	}
	delayWrite(&packet.ModalFormResponse{
		FormID: pk.FormID,
		ResponseData: protocol.Option(data),
	})
}

func (c *Connection) handleInventorySlot(pk *packet.InventorySlot){
	c.state.Exec(func(q *game.Qx) {
		q.SetItemOnInvSlot(pk.WindowID, pk.Slot, pk.NewItem)
	})
}

func (c *Connection) handlePlayerList(pk *packet.PlayerList){
	c.entInWorld.handlePlayerList(pk)
}