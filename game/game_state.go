package game

import (
	"fmt"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/game/movements"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

// GameState holds per-session Minecraft world data used by movement and packet output.
// Qx should be used for most GameState opteriation just like the *world.World in dragonfly
type GameState struct {
	clientData         login.ClientData
	entityRuntimeID    uint64
    player             *playerState
    blockMap           *blockmap.BlockMap
    tickInputDataFlags protocol.Bitset
	nextTickInMove     *movements.InMovement
	movement           *movements.Movement
    queue              chan *queueTransition
    tick               uint
    closed             chan struct{}
}

func NewGameState(conn *minecraft.Conn) *GameState {
	gs := &GameState{
		clientData:  conn.ClientData(),
		player:      newPlayerState(conn),
		blockMap:    blockmap.NewBlockMap(conn),
		queue:       make(chan *queueTransition, 512),
		closed: 	 make(chan struct{}),
		tick:        0,
	}
	fmt.Printf("Rewind size: %v\n", conn.GameData().PlayerMovementSettings.RewindHistorySize)
	gs.resetFlags()
	gs.movement = movements.NewMovement(gs.blockMap)
	gs.nextTickInMove = &movements.InMovement{}
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
	gs.player.tick()
	gs.blockMap.UpdateChunkCentre(gs.player.Position)
	gs.blockMap.RefreshMapWithRenderDistance()
	gs.doMovement()
}

func (gs *GameState) doMovement(){
	now := time.Now()
	out := gs.movement.SimMovementWithFlag(gs.splitInMovement(), &gs.tickInputDataFlags)
	gs.nextTickInMove = &movements.InMovement{}
	out.CopyOutToIn(gs.nextTickInMove)
	gs.copyOutMovement(out)
	fmt.Printf("Movement on tick %d: {position: %v velocity: %v onGround: %v}\n", gs.GStick(), gs.player.Position, gs.player.Velocity, gs.player.OnGround)
	fmt.Printf("Block pos based on pPos: %v\n", cube.PosFromVec3(utils.Mgl32Vec3Tomgl64Vec3(gs.player.Position)))
	fmt.Printf("Time used for tick %d: %0.3fms\n\n", gs.GStick(), time.Since(now).Seconds()*1000)
}

func (gs *GameState) splitInMovement() *movements.InMovement{
	in := gs.nextTickInMove
	in.Position = gs.player.Position
	in.OnGround = gs.player.OnGround
	in.Velocity = gs.player.Velocity
	in.Yaw = gs.player.Yaw
	return in
}

func (gs *GameState) copyOutMovement(out movements.OutMovement){
	ps := gs.player
	ps.Yaw = out.Yaw
	ps.Position = out.Position
	ps.Velocity = out.Velocity
	ps.OnGround = out.OnGround
}

func (gs *GameState) GStick() uint {
	return gs.tick
}
