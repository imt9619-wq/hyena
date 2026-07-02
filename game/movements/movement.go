package movements

import (
	"math"

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
	SneakProbeBBoxShrinks  = float64(0.025)
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
	m.stopOnEdge()
	m.simCollision()
	m.setOnGround()
	return m.splitOutMovement()
}

func (m *Movement) stopOnEdge(){
	tinyBBox := utils.TinyBBoxOnBBoxFace(utils.PlayerBBox(m.position), cube.FaceDown)
	if !(m.IsSneak() && utils.BBoxIntersectsSolid(m.world, tinyBBox) && m.velocity[1] <= 0){
		return
	}
	m.velocity[1] = 0
	probeOnEdge := func(axis int){
		if axis == 1 || m.velocity[axis] == 0{
			return
		} 
		planeSign := math.Abs(m.velocity[axis])/m.velocity[axis]
		planeFinal := m.velocity[axis]
		planeFinal -= planeSign*0.05
		if planeFinal != 0{
			if math.Abs(planeFinal)/planeFinal != planeSign{
				planeFinal = 0
			}
		}
		m.velocity[axis] = planeFinal
	}
	probeBBox := utils.BBoxOnBBoxFaceWithThreshold(utils.PlayerBBox(m.position).Grow(-SneakProbeBBoxShrinks), 
	cube.FaceDown, 
	MaxStepHeight+utils.ProbeOffset+SneakProbeBBoxShrinks)
	for i := range 3{
		for m.velocity[(i*2)%4] != 0 && m.velocity[i+(i%2)] != 0{
			deltas := m.velocity
			deltas[(i+2)%3] = 0
			if utils.BBoxIntersectsSolid(m.world, probeBBox.Translate(deltas)){
				break
			}
			probeOnEdge(i)
			probeOnEdge((i+1)%3)
		}
	}
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