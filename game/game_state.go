package game

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// GameState holds per-session Minecraft world data used by movement and packet output.
// Qx should be used for most GameState opteriation just like the *world.World in dragonfly
type GameState struct {
	conn            *minecraft.Conn
	entityRuntimeID uint64
	player          *playerState
	blockMap        *blockmap.BlockMap
	queue           chan *queueTransition
	closed          chan struct{}
}

func NewGameState(conn *minecraft.Conn) *GameState {
	gs := &GameState{
		conn:        conn,
		player:      newPlayerState(conn),
		blockMap:    blockmap.NewBlockMap(conn),
		queue:       make(chan *queueTransition, 512),
		closed: 	 make(chan struct{}),
	}
	gs.entityRuntimeID = conn.GameData().EntityRuntimeID
	gs.startRunningQueue()
	return gs
}

func (gs *GameState) UpdateRenderedChunks() {
	gs.blockMap.UpdateChunkCentre(gs.player.Position)
}

// return a pointer to PlayerAuthInput packet where the fields are filled out based on the 
// current GameState
func (gs *GameState) PlayerAuthInputWithState() *packet.PlayerAuthInput {
	ps := gs.player
	pk := &packet.PlayerAuthInput{
		Pitch:      ps.Pitch,
		Yaw:        ps.Yaw,
		Position:   ps.Position,
		MoveVector: mgl32.Vec2{ps.Velocity[0], ps.Velocity[2]},
	}
	return pk
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