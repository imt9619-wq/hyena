package handler

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

type movementAction interface {
	apply(*playerState, map[movementAction]struct{})
}

type movement struct {
	state          *gameState
	activeActions  map[movementAction]struct{}
}

func newMovement(state *gameState) *movement {
	return &movement{
		state:         state,
		activeActions: make(map[movementAction]struct{}, 3),
	}
}

func (m *movement) tick() {
	defer m.state.player.Unlock()
	m.state.player.Lock()

	for action := range m.activeActions {
		action.apply(m.state.player, m.activeActions)
	}
	m.applyVelocity()
	//m.checkCollision()
}

func (m *movement) checkCollision(){

}

func (m *movement) applyVelocity() {
	gravity := float32(-0.08)
	drag := float32(0.98)
	ps := m.state.player
	if !ps.onGround{
		ps.velocity[1] = (ps.velocity[1] + gravity) * drag
	}
	m.state.player.position.Add(m.state.player.velocity)
}

func (m *movement) startRunning() {
	m.activeActions[runAction{}] = struct{}{}
}

func (m *movement) stopRunning() {
	delete(m.activeActions, runAction{})
}

func (m *movement) startJumping() {
	m.activeActions[jumpAction{}] = struct{}{}
}

func (m *movement) stopJumping() {
	delete(m.activeActions, jumpAction{})
}

type runAction struct{}

func (runAction) apply(ps *playerState, active map[movementAction]struct{}) {
	slipperiness := float32(0.6)
	movementMult := float32(1.3)
	effectsMult := float32(1)

	jumpBoost := float32(0.2)
	if _, jumping := active[jumpAction{}]; !jumping {
		jumpBoost = 0
	}

	yawRad := float64(ps.yaw) * (math.Pi / 180)
	xVel := ps.velocity[0]
	zVel := ps.velocity[2]
	speed := xzSpeed(ps.velocity)

	momentum := speed * slipperiness * 0.91
	acceleration := float32(0.1) * movementMult * effectsMult * float32(math.Pow(0.6/float64(slipperiness), 3))
	if !ps.onGround {
		acceleration = 0
	}

	sinD := float32(0)
	cosD := float32(1)
	if speed > 0.003 {
		sinD = xVel / speed
		cosD = zVel / speed
	}
	ps.velocity[0] = momentum + acceleration*sinD + jumpBoost*float32(math.Sin(yawRad))
	ps.velocity[2] = momentum + acceleration*cosD + jumpBoost*float32(math.Cos(yawRad))
}

type jumpAction struct{}

func (jumpAction) apply(ps *playerState, active map[movementAction]struct{}) {
	jumpSpeed := float32(0.42)
	if ps.onGround {
		ps.position[1] = jumpSpeed
	}
}

func rotationToPitchAndYaw(r mgl32.Vec3) (yaw, pitch float32) {
	xz := math.Sqrt(math.Pow(float64(r[0]), 2) + math.Pow(float64(r[2]), 2))
	mag := math.Cbrt(math.Pow(xz, 2) + math.Pow(float64(r[1]), 2))

	pitch, yaw = float32(18/math.Pi), float32(18/math.Pi)
	if xz > 0.003 {
		yaw = float32(math.Acos(float64(r[2])/xz) * 180 / math.Pi)
	}
	if mag > 0.003 {
		pitch = float32(math.Acos(xz/mag) * 180 / math.Pi)
	}
	return
}

func xzSpeed(v mgl32.Vec3) float32 {
	return float32(math.Sqrt(math.Pow(float64(v[0]), 2) + math.Pow(float64(v[2]), 2)))
}
