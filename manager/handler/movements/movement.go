package movements

import (
	"fmt"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game"
	"github.com/imt9619-wq/hyena/game/physics"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type Movement struct {
	state        *game.GameState
    position     mgl64.Vec3
    velocity     mgl64.Vec3
    onGround     bool
    isrunning    bool
    isjumping    bool

    stateInWorld *physics.StateInWorld
}

func NewMovement(state *game.GameState) *Movement {
	return &Movement{
		state:    state,
		stateInWorld: physics.NewStateInWorld(state.BlockMap()),
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
	m.simCollision()
	m.setOnGround()
	m.pasteToPlayerState()
	fmt.Printf("Offset on tick %d: %+v\n", m.state.GStick(), m.stateInWorld.ScratchOffset())
	fmt.Printf("Movement on tick %d: {position: %v velocity: %v onGrond: %v}\n", m.state.GStick(), m.position, m.velocity, m.onGround)
	fmt.Printf("Time used for tick %d: %0.2fms\n", m.state.GStick(), time.Since(now).Seconds()*1000)
	fmt.Printf("Block pos based on pPos: %v\n\n", cube.PosFromVec3(m.position))
}

func (m *Movement) pasteToPlayerState() {
	ps := m.state.Player()
	ps.Velocity = utils.Mgl64Vec3Tomgl32Vec3(m.velocity)
	ps.Position = utils.Mgl64Vec3Tomgl32Vec3(m.position.Add(mgl64.Vec3{0, utils.NetworkOffset, 0}))
	ps.OnGround = m.onGround 
}

func (m *Movement) copyPlayerState() {
	ps := m.state.Player()
	m.velocity = utils.Mgl32Vec3Tomgl64Vec3(ps.Velocity)
	m.position = utils.Mgl32Vec3Tomgl64Vec3(ps.Position).Sub(mgl64.Vec3{0, utils.NetworkOffset, 0})

	m.onGround = ps.OnGround
}

func (m *Movement) setOnGround() {
	m.onGround= false
	halfW := utils.PlayerWidth / 2
	pos := m.position
	tinyBBox := cube.Box(
		pos[0]-halfW,
		pos[1]-utils.GroundProbeOffset,
		pos[2]-halfW,
		pos[0]+halfW,
		pos[1],
		pos[2]+halfW,
	)
	if m.velocity[1] == 0 && utils.BBoxIntersectsSolid(m.state.BlockMap(), tinyBBox) {
		m.onGround = true
		m.state.Player().SetFlag(packet.InputFlagVerticalCollision)
	}
}