package movements

import (
	"fmt"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game"
	"github.com/imt9619-wq/hyena/game/movements/physics"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const(
	AirborneSlipperiness   = float64(1)
	SlipperinessToFriction = float64(0.91)
	SprintMovementMult     = float64(1.3)
	SprintJumpBoost        = float64(0.2)
	JumpSpeed              = float64(0.42)
	MomentumThreshold      = float64(0.003)
	MaxStepHeight          = float64(0.6)
	ClimbSpeed             = float64(0.1176)
)

type Movement struct {
	state        *game.GameState
    position     mgl64.Vec3
    velocity     mgl64.Vec3
    onGround     bool
    isrunning    bool
    isjumping    bool
	onClimb      bool
	slipperiness float64

    stateInWorld *physics.StateInWorld
}

func NewMovement(state *game.GameState) *Movement {
	return &Movement{
		state:    state,
		stateInWorld: physics.NewStateInWorld(state.BlockMap()),
	}
}

func (m *Movement) Tick() {
	now := time.Now()
	m.copyPlayerState()
	m.doMotions()
	m.simCollision()
	m.setOnGround()
	m.pasteToPlayerState()
	fmt.Printf("Offset on tick %d: %+v\n", m.state.GStick(), m.stateInWorld.ScratchOffset())
	fmt.Printf("Movement on tick %d: {position: %v velocity: %v onGround: %v}\n", m.state.GStick(), m.position, m.velocity, m.onGround)
	fmt.Printf("Block pos based on pPos: %v\n", cube.PosFromVec3(m.position))
	fmt.Printf("Time used for tick %d: %0.3fms\n\n", m.state.GStick(), time.Since(now).Seconds()*1000)
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
	m.position = utils.RoundVecTo5Decimal(m.position)

	m.onGround = ps.OnGround
}

func (m *Movement) setOnGround() {
	m.onGround= false
	tinyBBox := utils.TinyBBoxOnBBoxFace(utils.PlayerBBox(m.position), cube.FaceDown)
	if m.velocity[1] == 0 && utils.BBoxIntersectsSolid(m.state.BlockMap(), tinyBBox) {
		m.onGround = true
		m.state.SetFlag(packet.InputFlagVerticalCollision)
	}
}
