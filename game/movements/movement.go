package movements

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/game/input"
	"github.com/imt9619-wq/hyena/utils"
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

type MovementFlags struct {
    OnGround            bool
    OnClimb             bool
    HorizontalCollision bool
    VerticalCollision   bool
    StartedJumping      bool
    WantUp              bool
    WantDown            bool
}

type Movement struct {
	input.Inputs
	world        *blockmap.BlockMap
    position     mgl64.Vec3
    velocity     mgl64.Vec3
    yaw          float64
    onGround     bool
    onClimb      bool
    slipperiness float64
	baseSpeed    float64
	jumpCooldown int

    flag         MovementFlags
}

func SimMovementsInWorld(in *InMovement, world *blockmap.BlockMap) *OutMovement{
	m := &Movement{world: world}
	m.copyInMovement(in)
	m.doMotions()
	m.stopOnEdge()
	m.simCollision()
	m.setOnGround()
	out := m.splitOutMovement()
	m = nil
	return out
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

func (m *Movement) setOnGround() {
	m.onGround = false
	tinyBBox := utils.TinyBBoxOnBBoxFace(utils.PlayerBBox(m.position), cube.FaceDown)
	if m.velocity[1] == 0 && utils.BBoxIntersectsSolid(m.world, tinyBBox) {
		m.onGround = true
		m.flag.VerticalCollision = true
	}
}

func (m *Movement) Velocity() mgl64.Vec3{
	return utils.RoundVecTo5Decimal(m.velocity)
}

func (m *Movement) Position() mgl64.Vec3{
	return utils.RemoveDeltaEpsilon(m.position)
}

func (m *Movement) BBox() cube.BBox{
	return m.bbox(m.Position())
}

func (m *Movement) bbox(pos mgl64.Vec3) cube.BBox{
	if m.IsSneak(){
		return utils.PlayerSneakBBox(pos)
	}
	return utils.PlayerBBox(pos)
}

func (m *Movement) World() utils.BlockSourse{
	return m.world
}