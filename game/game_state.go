package game

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// GameState holds per-session Minecraft world data used by movement and packet output.
// Qx should be used for most GameState opteriation just like the *world.World in dragonfly
type GameState struct {
	entityRuntimeID    uint64
    player             *playerState
    blockMap           *blockmap.BlockMap
    tickInputDataFlags protocol.Bitset
    queue              chan *queueTransition
    tick               uint64
    closed             chan struct{}
}

func NewGameState(conn *minecraft.Conn) *GameState {
	gs := &GameState{
		player:      newPlayerState(conn),
		blockMap:    blockmap.NewBlockMap(conn),
		tickInputDataFlags: protocol.NewBitset(packet.PlayerAuthInputBitsetSize),
		queue:       make(chan *queueTransition, 512),
		closed: 	 make(chan struct{}),
		tick:        0,
	}
	gs.entityRuntimeID = conn.GameData().EntityRuntimeID
	gs.startRunningQueue()
	return gs
}

// close the qx queue loop, will panic if close again after closing
func (gs *GameState) Close() {
	close(gs.closed)
}

func (gs *GameState) BlockMap() *blockmap.BlockMap {
	return gs.blockMap
}

func (gs *GameState) EntityRunTimeId() uint64 {
	return gs.entityRuntimeID
}

func (gs *GameState) Tick() {
	gs.tick++
	gs.resetFlags()
	gs.setInputFlagBlockBreakingDelayEnabled()
	gs.player.tick()
	gs.blockMap.UpdateChunkCentre(gs.player.Position)
	gs.blockMap.RefreshMapWithRenderDistance()
}

func (gs *GameState) GStick() uint64 {
	return gs.tick
}

// return a pointer to PlayerAuthInput packet where the fields are filled out based on the 
// current GameState
func (gs *GameState) PlayerAuthInputWithState() *packet.PlayerAuthInput {
	pk := &packet.PlayerAuthInput{}
	pk.Tick = gs.tick
	pk.InputMode = packet.InputModeTouch
	pk.PlayMode = packet.PlayModeTeaser
	pk.InteractionModel = packet.InteractionModelTouch
	pk.BlockActions = nil
	pk.InputData = gs.tickInputDataFlags
	pk.ItemInteractionData = protocol.UseItemTransactionData{}
	pk.ItemStackRequest = protocol.ItemStackRequest{}
	pk.VehicleRotation = mgl32.Vec2{}
	pk.ClientPredictedVehicle = 0
	pk.AnalogueMoveVector = mgl32.Vec2{}
	pk.CameraOrientation = mgl32.Vec3{}
	gs.Player().setPlayerAuthInputWithPlayerState(pk)
	return pk
}

func (gs *GameState) SetFlag(flag int){
	gs.tickInputDataFlags.Set(flag)
}

// Reset all bits in ps.tickInputDataFlags to 0
func (gs *GameState) resetFlags() {
	inputDataFlags := gs.tickInputDataFlags
	for i := range inputDataFlags.Len(){
		inputDataFlags.Unset(i)
	}
}

func (gs *GameState) setInputFlagBlockBreakingDelayEnabled() {
	gs.SetFlag(packet.InputFlagBlockBreakingDelayEnabled)
}