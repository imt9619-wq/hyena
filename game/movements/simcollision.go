package movements

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game/physics"
	"github.com/imt9619-wq/hyena/utils"
)

func (m *Movement) simCollision(){
	out := physics.EntityCollision(m)
	pos, velocity := m.doStepAssist(out)
	for axis, plane := range velocity{
		if axis != 1 && plane == 0 && m.velocity[axis] != 0{
			m.flag.HorizontalCollision = true
			break
		}
	}
	m.pasteStateToMovements(pos, velocity)
}

func (m *Movement) doStepAssist(op physics.OutPhyState) (pos, velocity mgl64.Vec3){
	pos, velocity = op.Position, op.Velocity
	if !m.onGround || utils.DeltaIsZero(m.velocity) || m.onClimb{
		return
	}

	var stepHeight float64 = 0
	var ceilHeight float64 = 1000
	var walkStairVelocity mgl64.Vec3
	for axis, plane := range velocity{
		if plane == 0 && m.velocity[axis] != 0 && axis != 1{
			walkStairVelocity[axis] = m.velocity[axis]
		}else{
			walkStairVelocity[axis] = 0
		}
	}
	var veloLen float64 = 1
	zeroVelo := mgl64.Vec3{}
	if walkStairVelocity != zeroVelo{
		veloLen = walkStairVelocity.Len()
	}
	pBBoxInStairs := m.bboxFunc(op.Position)
	pBBoxInStairs = pBBoxInStairs.Extend(walkStairVelocity.Mul(utils.ProbeOffset/veloLen))
	pBBoxInStairs = pBBoxInStairs.ExtendTowards(cube.FaceUp, MaxStepHeight)

	for _, blockBox := range utils.SweptBBoxInBBox(pBBoxInStairs, m.world){
		if pBBoxInStairs.IntersectsWith(blockBox){
			if blockBox.Min()[1] >= m.BBox().Max()[1]{
				ceilHeight = min(ceilHeight, blockBox.Min()[1]-m.BBox().Max()[1])
			}else if blockBox.Max()[1] >= m.BBox().Min()[1]{
				stepHeight = max(stepHeight, blockBox.Max()[1]-m.BBox().Min()[1])
			}				
		}
	}
	if stepHeight > MaxStepHeight || stepHeight == 0 || ceilHeight < stepHeight{
		return
	}
	// jump cancel
	velocityAfterStair := m.velocity
	if m.Space.Pressed && m.velocity[1] == JumpSpeed && stepHeight >= JumpSpeed{
		velocityAfterStair[1] = 0
	}
	
	stepEnt := physics.NopEntity{
		Pos: m.position.Add(mgl64.Vec3{0, stepHeight, 0}),
		Vec: velocityAfterStair,
		Bs: m.world,
	}
	stepEnt.AAbb = m.bbox(stepEnt.Position())
	stepOp := physics.EntityCollision(stepEnt)
	if stepOp.Position.Sub(m.position).Len() <= pos.Sub(m.position).Len(){
		return
	}
	return stepOp.Position, stepOp.Velocity
}

func (m *Movement) pasteStateToMovements(pos, velocity mgl64.Vec3){
	m.velocity = velocity
	m.position = pos
	if mgl64.FloatEqualThreshold(m.position[1], float64(m.world.Dimension().Range()[0]), utils.Negligible){
		m.position[1] = float64(m.world.Dimension().Range()[0])
	}
}