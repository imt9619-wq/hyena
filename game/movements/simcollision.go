package movements

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game/movements/physics"
	"github.com/imt9619-wq/hyena/utils"
)

func (m *Movement) simCollision(){
	pos, velocity := m.doNormalCollisionThenStepAssist()
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
	pBBoxInStairs := utils.PlayerBBox(op.Position)
	pBBoxInStairs = pBBoxInStairs.Extend(walkStairVelocity.Mul(utils.ProbeOffset/walkStairVelocity.Len()))
	pBBoxInStairs = pBBoxInStairs.ExtendTowards(cube.FaceUp, MaxStepHeight)

	for _, blockBox := range utils.SweptBBoxInBBox(pBBoxInStairs, m.state.BlockMap()){
		if pBBoxInStairs.IntersectsWith(blockBox){
			if blockBox.Min()[1] >= op.AABB.Max()[1]{
				ceilHeight = min(ceilHeight, blockBox.Min()[1]-op.AABB.Max()[1])
			}else if blockBox.Max()[1] >= op.AABB.Min()[1]{
				stepHeight = max(stepHeight, blockBox.Max()[1]-op.AABB.Min()[1])
			}				
		}
	}
	if stepHeight > MaxStepHeight || stepHeight == 0 || ceilHeight < stepHeight{
		return
	}
	// jump cancel
	velocityAfterStair := m.velocity
	if m.isjumping && m.velocity[1] == JumpSpeed && stepHeight >= JumpSpeed{
		velocityAfterStair[1] = 0
	}
	
	stepOp := m.simAState(m.position.Add(mgl64.Vec3{0, stepHeight, 0}), velocityAfterStair)
	if stepOp.Position.Sub(m.position).Len() <= pos.Sub(m.position).Len(){
		return
	}
	return stepOp.Position, stepOp.Velocity
}

func (m *Movement) pasteStateToMovements(pos, velocity mgl64.Vec3){
	m.velocity = velocity
	m.position = pos
	if mgl64.FloatEqualThreshold(m.position[1], float64(m.state.BlockMap().Dimension().Range()[0]), utils.Negligible){
		m.position[1] = float64(m.state.BlockMap().Dimension().Range()[0])
	}
}

func (m *Movement) doNormalCollisionThenStepAssist() (mgl64.Vec3, mgl64.Vec3){
	return m.doStepAssist(m.simAState(m.position, m.velocity))
}

func (m *Movement) simAState(pos, velocity mgl64.Vec3) physics.OutPhyState{
	out := m.stateInWorld.SimState(physics.InPhyState{
		Position: pos,
		Velocity: velocity,
		BBoxFunc: utils.PlayerBBox,
	})
	return out
}