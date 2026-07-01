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
	AirborneAccelration    = float64(0.026)
	SlipperinessToFriction = float64(0.91)
	SprintMovementMult     = float64(1.3)
	SprintJumpBoost        = float64(0.2)
	JumpSpeed              = float64(0.42)
	MomentumThreshold      = float64(0.003)
	MaxStepHeight          = float64(0.6)
	ClimbSpeed             = float64(0.1176)
	DefaultBaseSpeed       = float64(0.1)
)

type Movement struct {
	Inputs
	world        *blockmap.BlockMap
    position     mgl64.Vec3
    velocity     mgl64.Vec3
    yaw          float64
    onGround     bool
    onClimb      bool
    slipperiness float64
	baseSpeed    float64
	bboxFunc   utils.BBoxFunc

	blockSource  utils.BlockSourse
    flag         *protocol.Bitset
    stateInWorld *physics.StateInWorld
}

func NewMovement(world *blockmap.BlockMap) *Movement {
	return &Movement{
		world: world,
		stateInWorld: physics.NewStateInWorld(),
		baseSpeed: DefaultBaseSpeed,
	}
}

func (m *Movement) SimMovements(in *InMovement) *OutMovement{
	m.copyInMovement(in)
	m.doMotions()
	m.simCollision()
	m.setOnGround()
	return m.splitOutMovement()
}

func (m *Movement) SimMovementsWithFlags(in *InMovement) *OutMovement{
	m.flag = in.Input.InputFlags
	return m.SimMovements(in)
}

func (m *Movement) setOnGround() {
	m.onGround = false
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