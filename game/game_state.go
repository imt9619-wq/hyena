package game

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/game/input"
	"github.com/imt9619-wq/hyena/game/itemstack"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/imt9619-wq/hyena/utils/pkbuf"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// GameState holds per-session Minecraft world data used by movement and packet output.
// Qx should be used for most GameState opteriation just like the *world.World in dragonfly
type GameState struct {
	clientData      *login.ClientData
    entityRuntimeID uint64
    blockMap        *blockmap.BlockMap
    player          *playerState
	items			*itemstack.PlayerItemStack

    tickInputDataFlags protocol.Bitset
    in                 input.Inputs
    moveBuf            *moveBuf

    queue    chan *queueTransition
    currTick uint
    closed   chan struct{}

    packets *pkbuf.PacketBuffer
}

func NewGameState(conn *minecraft.Conn) *GameState {
	gs := &GameState{}
	gs.packets = pkbuf.NewPacketBuffer(10)
	gs.entityRuntimeID = conn.GameData().EntityRuntimeID
	gs.blockMap = blockmap.NewBlockMap(conn, gs.packets)
	gs.moveBuf = newMoveBuf(conn)
	gs.queue = make(chan *queueTransition, 512)
	gs.closed = make(chan struct{})
	data := conn.ClientData()
	gs.clientData = &data
	gs.resetFlags()
	gs.player = newPlayerState(conn, movements.NewMovement(gs.blockMap))
	gs.items = itemstack.NewPlayerItemStack(conn, gs.packets)
	gs.startRunningQueue()
	return gs
}

// close the qx queue loop, will panic if close again after closing
func (gs *GameState) Close() {
	close(gs.closed)
}

func (gs *GameState) tick() {
	gs.currTick++
	gs.setInputFlagBlockBreakingDelayEnabled()
	gs.blockMap.UpdateChunkCentre(gs.player.position)
	gs.blockMap.RefreshMapWithRenderDistance()
	gs.blockMap.RequestSubChunkInQuery()
	gs.handleInput()
	gs.moveTick()
	gs.tickReset()
}

func (gs *GameState) handleInput(){
	if gs.in.RightClick.Pressed{
		itemData := &protocol.UseItemTransactionData{
			Position: gs.player.position,
			TriggerType: protocol.TriggerTypePlayerInput,
			BlockFace: 255,
			HotBarSlot: int32(gs.items.HeldSlot()),
			ActionType: protocol.UseItemActionClickAir,
			ClickedPosition: mgl32.Vec3{0.5, 0.5, 0.5},
			HeldItem: gs.items.SlotInstance(gs.items.HeldSlot()),
		}
		gs.packets.Append(&packet.InventoryTransaction{
			TransactionData: itemData,
		})
	}
}

func (gs *GameState) tickReset(){
	gs.resetFlags()
	gs.in = gs.in.NextTickPresses()
}

func (gs *GameState) EntityRunTimeId() uint64 {
	return gs.entityRuntimeID
}

func (gs *GameState) GStick() uint {
	return gs.currTick
}

func (gs *GameState) PacketBuf() *pkbuf.PacketBuffer{
	return gs.packets
}