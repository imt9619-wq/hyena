package game

import (
	"fmt"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// GameState holds per-session Minecraft world data used by movement and packet output.
// Qx should be used for most GameState opteriation just like the *world.World in dragonfly
type GameState struct {
	clientData         login.ClientData
    entityRuntimeID    uint64
    blockMap           *blockmap.BlockMap
    tickInputDataFlags protocol.Bitset

    player   *playerState
    movement *movements.Movement
    moveBuf  *moveBuf

    queue  chan *queueTransition
    tick   uint
    closed chan struct{}
}

func NewGameState(conn *minecraft.Conn) *GameState {
	gs := &GameState{
		clientData:  conn.ClientData(),
		player:      newPlayerState(conn),
		blockMap:    blockmap.NewBlockMap(conn),
		moveBuf:     newMoveBuf(conn),
		queue:       make(chan *queueTransition, 512),
		closed: 	 make(chan struct{}),
		tick:        0,
	}
	fmt.Printf("Rewind size: %v\n", conn.GameData().PlayerMovementSettings.RewindHistorySize)
	gs.resetFlags()
	gs.movement = movements.NewMovement(gs.blockMap)
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
	gs.setInputFlagBlockBreakingDelayEnabled()
	gs.blockMap.UpdateChunkCentre(gs.player.Position)
	gs.blockMap.RefreshMapWithRenderDistance()
	gs.doMovement()
	gs.tickReset()
}

func (gs *GameState) doMovement(){
	now := time.Now()
	out := gs.movement.SimMovementsWithFlags(gs.player.splitInMovement(&gs.tickInputDataFlags))
	gs.setStateChangeFlags(out)	
	gs.player.copyOutMovement(out)
	gs.moveBuf.addTick(out)
	gs.player.in = out.Input.NextTickPresses()
	fmt.Printf("Movement on tick %d: {position: %v velocity: %v onGround: %v}\n", gs.GStick(), gs.player.Position.Sub(mgl32.Vec3{0, float32(utils.NetworkOffset)}), gs.player.Velocity, gs.player.OnGround)
	fmt.Printf("Block pos based on pPos: %v\n", cube.PosFromVec3(utils.Mgl32Vec3Tomgl64Vec3(gs.player.Position)))
	fmt.Printf("Time used for tick %d: %0.3fms\n\n", gs.GStick(), time.Since(now).Seconds()*1000)
}

func (gs *GameState) GStick() uint {
	return gs.tick
}

func (gs *GameState) Inputs() *movements.Inputs{
	return &gs.player.in
}

func (gs *GameState) Player() *playerState {
	return gs.player
}

func (gs *GameState) setStateChangeFlags(nowOut *movements.OutMovement){
	out, ok := gs.moveBuf.outMoveWithTick(gs.tick-1)
	nowIn := nowOut.Input
	lastIn := movements.Inputs{}
	if ok{
		lastIn = out.Input
	}
	if !lastIn.Space.Pressed && nowIn.Space.Pressed{
		gs.SetFlag(packet.InputFlagJumpPressedRaw)
	}
	if lastIn.Space.Pressed && !nowIn.Space.Pressed{
		gs.SetFlag(packet.InputFlagJumpReleasedRaw)
	}
	if !lastIn.Shift.Pressed && nowIn.Shift.Pressed{
		gs.SetFlag(packet.InputFlagSneakPressedRaw)
		gs.SetFlag(packet.InputFlagStartSneaking)
	}
	if lastIn.Shift.Pressed && !nowIn.Shift.Pressed{
		gs.SetFlag(packet.InputFlagStopSneaking)
		gs.SetFlag(packet.InputFlagSneakReleasedRaw)
	}
	if !ok{
		return
	}
	if lastIn.InputFlags.Load(packet.InputFlagSprinting) && !nowIn.InputFlags.Load(packet.InputFlagSprinting){
		gs.SetFlag(packet.InputFlagStopSprinting)
	}
}

func flagLoad(flags *protocol.Bitset, flag int) bool{
	if flags == nil{
		return false
	}
	return (*flags).Load(flag)
}

func (gs *GameState) tickReset(){
	gs.resetFlags()
	gs.player.in.ServerSpeedAdd = mgl32.Vec3{}
	gs.player.in.InputFlags = &gs.tickInputDataFlags
}