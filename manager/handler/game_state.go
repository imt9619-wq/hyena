package handler

import (
	"sync/atomic"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/manager/handler/blockmap"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// gameState holds per-session Minecraft world data used by movement and packet output.
// Qx is only used for calling blockmap methods currently
type gameState struct {
	conn            *minecraft.Conn
	entityRuntimeID uint64
	player          *playerState
	flushedTick     *atomic.Int32
	blockMap        *blockmap.BlockMap
	queue           chan *queueTransition
}

func newGameState(conn *minecraft.Conn) *gameState {
	gs := &gameState{
		conn:        conn,
		player:      newPlayerState(conn),
		flushedTick: &atomic.Int32{},
		blockMap:    blockmap.NewBlockMap(conn),
		queue:       make(chan *queueTransition, 512),
	}
	gs.entityRuntimeID = conn.GameData().EntityRuntimeID
	return gs
}

func (gs *gameState) flush() {
	defer gs.flushedTick.Add(1)
	// gs.writePlayerAuthInput()
	gs.blockMap.UpdateChunkCentre(gs.player.position)
}

func (gs *gameState) writePlayerAuthInput() {
	ps := gs.player
	gs.conn.WritePacket(&packet.PlayerAuthInput{
		Pitch:      ps.pitch,
		Yaw:        ps.yaw,
		Position:   ps.position,
		MoveVector: mgl32.Vec2{ps.velocity[0], ps.velocity[2]},
	})
}
