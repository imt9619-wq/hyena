package movements

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/game/movements/physics"
	"github.com/imt9619-wq/hyena/utils"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
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
	world        *blockmap.BlockMap
    position     mgl64.Vec3
    velocity     mgl64.Vec3
    yaw          float64
    onGround     bool
    isrunning    bool
    isjumping    bool
    onClimb      bool
    slipperiness float64

    flag         *protocol.Bitset
    stateInWorld *physics.StateInWorld
}

func NewMovement(world *blockmap.BlockMap) *Movement {
	return &Movement{
		world: world,
		stateInWorld: physics.NewStateInWorld(world),
	}
}

func (m *Movement) SimMovementWithFlag(in *InMovement, flag *protocol.Bitset) *OutMovement{
	m.flag = flag
	m.copyInMovement(in)
	m.doMotions()
	m.simCollision()
	m.setOnGround()
	return m.splitOutMovement()
}

func (m *Movement) SimMovement(in *InMovement) *OutMovement{
	return m.SimMovementWithFlag(in, nil)
}

func (m *Movement) setOnGround() {
	m.onGround= false
	tinyBBox := utils.TinyBBoxOnBBoxFace(utils.PlayerBBox(m.position), cube.FaceDown)
	if m.velocity[1] == 0 && utils.BBoxIntersectsSolid(m.world, tinyBBox) {
		m.onGround = true
		m.setFlag(packet.InputFlagVerticalCollision)
	}
}

func (m *Movement) setFlag(i int){
	if m.flag == nil{
		return
	}
	(*m.flag).Set(i)
}