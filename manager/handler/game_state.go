package handler

import (
	"sync/atomic"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/imt9619-wq/hyena/manager/sim"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// gameState holds per-session state and the serialised mutation queue.
type gameState struct {
	conn            *minecraft.Conn
	entityRuntimeID uint64
	session         *sim.Session
	flushedTick     *atomic.Int32
	queue           chan *queueTransition
}

func newGameState(conn *minecraft.Conn) *gameState {
	gs := &gameState{
		conn:        conn,
		session:     sim.NewSession(conn),
		flushedTick: &atomic.Int32{},
		queue:       make(chan *queueTransition, 512),
	}
	gs.entityRuntimeID = conn.GameData().EntityRuntimeID
	return gs
}

func (gs *gameState) flush() {
	defer gs.flushedTick.Add(1)
	// gs.writePlayerAuthInput()
	gs.session.BlockMap.UpdateChunkCentre(gs.session.Player.Position)
}

func (gs *gameState) writePlayerAuthInput() {
	p := gs.session.Player
	gs.conn.WritePacket(&packet.PlayerAuthInput{
		Pitch:      p.Pitch,
		Yaw:        p.Yaw,
		Position:   p.Position,
		MoveVector: mgl32.Vec2{p.Velocity[0], p.Velocity[2]},
	})
}
