package game

import (
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// GameState holds per-session Minecraft world data used by movement and packet output.
// Qx should be used for most GameState opteriation just like the *world.World in dragonfly
type GameState struct {
	entityRuntimeID uint64
	player          *playerState
	blockMap        *blockmap.BlockMap
	queue           chan *queueTransition
	tick			uint64
	closed          chan struct{}
}

func NewGameState(conn *minecraft.Conn) *GameState {
	gs := &GameState{
		player:      newPlayerState(conn),
		blockMap:    blockmap.NewBlockMap(conn),
		queue:       make(chan *queueTransition, 512),
		closed: 	 make(chan struct{}),
		tick:        0,
	}
	gs.entityRuntimeID = conn.GameData().EntityRuntimeID
	gs.startRunningQueue()
	return gs
}

func (gs *GameState) UpdateRenderedChunks() {
	gs.blockMap.UpdateChunkCentre(gs.player.Position)
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
	gs.player.tick()
}

func (gs *GameState) GStick() uint64 {
	return gs.tick
}

// return a pointer to PlayerAuthInput packet where the fields are filled out based on the 
// current GameState
func (gs *GameState) PlayerAuthInputWithState() *packet.PlayerAuthInput {
	return gs.Player().playerAuthInputWithState(gs.tick)
}