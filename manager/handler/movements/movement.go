package movements

import (
	"fmt"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// networkOffset can be found at github.com\df-mc\dragonfly\server\player.(ptype).NetworkOffset()
const (
	playerHeight        = float64(1.8)
	playerWidth         = float64(0.6)
	defaultSlipperiness = float64(0.6)
	sprintMovementMult  = float64(1.3)
	sprintJumpBoost     = float64(0.2)
	jumpSpeed           = float64(0.42)
	networkOffset       = float64(1.62)
	momentumThreshold   = float64(0.003)
	groundProbeOffset   = float64(0.003)
	negligible          = float64(0.003)
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
	now := time.Now()
	m.copyPlayerState()
	m.doMotions()
	m.applyVelocity()
	m.applyCollision(m.getCollision())
	m.setOnGround()
	m.pasteToPlayerState()
	fmt.Printf("Movement on tick %d: {position: %v velocity: %v onGrond: %v}\n", m.state.GStick(), m.position, m.velocity, m.onGround)
	fmt.Printf("Time used for tick %d: %0.2fms\n", m.state.GStick(), time.Since(now).Seconds()*1000)
	fmt.Printf("Block pos based on pPos: %v\n\n", Mgl64Vec3ToCubePos(m.position))
}

func (m *Movement) pasteToPlayerState() {
	ps := m.state.Player()
	ps.Velocity = mgl64Vec3Tomgl32Vec3(m.velocity)
	ps.Position = mgl64Vec3Tomgl32Vec3(m.position.Add(mgl64.Vec3{0, networkOffset, 0}))
	ps.OnGround = m.onGround 
}

func (m *Movement) copyPlayerState() {
	ps := m.state.Player()
	m.velocity = mgl32Vec3Tomgl64Vec3(ps.Velocity)
	m.position = mgl32Vec3Tomgl64Vec3(ps.Position).Sub(mgl64.Vec3{0, networkOffset, 0})
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