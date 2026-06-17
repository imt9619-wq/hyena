package movements

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const (
	playerHeight        = float64(1.8)
	playerWidth         = float64(0.6)
	defaultSlipperiness = float64(0.6)
	sprintMovementMult  = float64(1.3)
	sprintJumpBoost     = float64(0.2)
	jumpSpeed           = float64(0.42)
	momentumThreshold   = float64(0.003)
	groundProbeOffset   = float64(0.003)
	negligible          = float64(0.0000003)
)

type Movement struct {
	state         *game.GameState
	position      mgl64.Vec3
	velocity      mgl64.Vec3
	onGround      bool
	isrunning     bool
	isjumping     bool

	scratch *collisionScratch
}

func NewMovement(state *game.GameState) *Movement {
	return &Movement{
		state:    state,
		scratch:  newCollisionScratch(),
	}
}

func (m *Movement) playerPosBeforeVelocityApply() mgl64.Vec3 {
	return m.position.Sub(m.velocity)
}

func (m *Movement) Tick() {
	m.copyPlayerState()
	m.doMotions()
	m.applyVelocity()
	m.applyCollision(m.getCollision())
	m.setOnGround()
	m.pasteToPlayerState()
	/*if m.state.GStick()%10 == 0{
		fmt.Printf("Movement on tick %d: %+v\n", m.state.GStick(), m)
	}*/
}

func (m *Movement) pasteToPlayerState() {
	ps := m.state.Player()
	ps.Velocity = mgl64Vec3Tomgl32Vec3(m.velocity)
	ps.Position = mgl64Vec3Tomgl32Vec3(m.position)  
	ps.OnGround = m.onGround 
}

func (m *Movement) copyPlayerState() {
	ps := m.state.Player()
	m.velocity = mgl32Vec3Tomgl64Vec3(ps.Velocity)
	m.position = mgl32Vec3Tomgl64Vec3(ps.Position)
	m.onGround = ps.OnGround
}

func (m *Movement) setOnGround() {
	m.onGround= false
	halfW := playerWidth / 2
	pos := m.position
	tinyBBox := cube.Box(
		pos[0]-halfW,
		pos[1]-groundProbeOffset,
		pos[2]-halfW,
		pos[0]+halfW,
		pos[1],
		pos[2]+halfW,
	)
	if m.velocity[1] == 0 && m.bboxIntersectsSolid(tinyBBox) {
		m.onGround = true
		m.state.Player().SetFlag(packet.InputFlagVerticalCollision)
	}
}