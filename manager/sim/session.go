package sim

import (
	"github.com/imt9619-wq/hyena/manager/handler/blockmap"
	"github.com/sandertv/gophertunnel/minecraft"
)

// Session holds per-connection simulation state: player, world blocks, and movement.
type Session struct {
	Player   *Player
	BlockMap *blockmap.BlockMap
	movement *Movement
}

// NewSession creates simulation state for a connected player.
func NewSession(conn *minecraft.Conn) *Session {
	s := &Session{
		Player:   NewPlayer(conn),
		BlockMap: blockmap.NewBlockMap(conn),
	}
	s.movement = newMovement(s)
	return s
}

// Tick advances movement simulation by one step.
func (s *Session) Tick() {
	s.movement.tick()
}

// SetRunning sets the sprint-forward input flag.
func (s *Session) SetRunning(running bool) {
	s.movement.isRunning = running
}

// SetJumping sets the jump input flag.
func (s *Session) SetJumping(jumping bool) {
	s.movement.isJumping = jumping
}
