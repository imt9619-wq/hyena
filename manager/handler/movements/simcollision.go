package movements

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/utils"
)

func (m *Movement) simCollision(){
	pos, velocity := m.doNormalCollisionThenStepAssist()
	m.pasteStateToMovements(pos, velocity)
}

func (m *Movement) doStepAssist() (pos, velocity mgl64.Vec3){
	pos, velocity = m.stateInWorld.Position, m.stateInWorld.Velocity
	if !m.onGround || utils.DeltaIsZero(m.velocity) || m.onClimb{
		return
	}

	sw :=  m.stateInWorld
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
	pBBoxInStairs := utils.PlayerBBox(sw.Position)
	pBBoxInStairs = pBBoxInStairs.Extend(walkStairVelocity.Mul(utils.ProbeOffset/walkStairVelocity.Len()))
	pBBoxInStairs = pBBoxInStairs.ExtendTowards(cube.FaceUp, MaxStepHeight)

	for _, blockBox := range utils.SweptBBoxInBBox(pBBoxInStairs, m.state.BlockMap()){
		if pBBoxInStairs.IntersectsWith(blockBox){
			if blockBox.Min()[1] >= sw.AABB.Max()[1]{
				ceilHeight = min(ceilHeight, blockBox.Min()[1]-sw.AABB.Max()[1])
			}else if blockBox.Max()[1] >= sw.AABB.Min()[1]{
				stepHeight = max(stepHeight, blockBox.Max()[1]-sw.AABB.Min()[1])
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
	
	stepPos, stepVelocity := m.simAState(m.position.Add(mgl64.Vec3{0, stepHeight, 0}), velocityAfterStair)
	if stepPos.Sub(m.position).Len() <= pos.Sub(m.position).Len(){
		return
	}
	return stepPos, stepVelocity
}

func (m *Movement) pasteStateToMovements(pos, velocity mgl64.Vec3){
	m.velocity = velocity
	m.position = pos
	if mgl64.FloatEqualThreshold(m.position[1], float64(m.state.BlockMap().Dimension().Range()[0]), utils.Negligible){
		m.position[1] = float64(m.state.BlockMap().Dimension().Range()[0])
	}
}

func (m *Movement) doNormalCollisionThenStepAssist() (mgl64.Vec3, mgl64.Vec3){
	m.simAState(m.position, m.velocity)
	return m.doStepAssist()
}

func (m *Movement) simAState(pos, velocity mgl64.Vec3) (newPos, newVelocity mgl64.Vec3){
	m.stateInWorld.Position = pos
	m.stateInWorld.Velocity = velocity
	m.stateInWorld.BBoxFunc = utils.PlayerBBox
	m.stateInWorld.SimState()
	return m.stateInWorld.Position, m.stateInWorld.Velocity
}