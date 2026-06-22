package movements

import (
	"fmt"

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
	if !m.onGround || m.velocity[1] != 0 || utils.DeltaIsZero(m.velocity){
		return
	}

	sw :=  m.stateInWorld
	var stepHeight float64 = 0
	var ceilHeight float64 = 1000
	pBBoxInStairs := utils.PlayerBBox(sw.Position)
	pBBoxInStairs = pBBoxInStairs.Extend(m.velocity.Mul(2*utils.HoriProbeOffset/m.velocity.Len()))
	pBBoxInStairs = pBBoxInStairs.ExtendTowards(cube.FaceUp, utils.MaxStepHeight)

	for pos, model := range utils.SweptModelsInBBox(pBBoxInStairs, m.state.BlockMap()){
		for _, blockBox := range utils.BBoxes(model, pos, m.state.BlockMap()){
			if pBBoxInStairs.IntersectsWith(blockBox){
				if blockBox.Min()[1] >= sw.AABB.Max()[1]{
					ceilHeight = min(ceilHeight, blockBox.Min()[1]-sw.AABB.Max()[1])
				}else if blockBox.Max()[1] >= sw.AABB.Min()[1]{
					stepHeight = max(stepHeight, blockBox.Max()[1]-sw.AABB.Min()[1])
				}				
			}
		}
	}
	fmt.Printf("ceilHeight: %v, stepHeight: %v\n", ceilHeight, stepHeight)
	if stepHeight > utils.MaxStepHeight || ceilHeight < stepHeight{
		return
	}
	stepPos, stepVelocity := m.simAState(m.position.Add(mgl64.Vec3{0, stepHeight, 0}), m.velocity)
	if stepPos.Sub(m.playerPosBeforeVelocityApply()).Len() <= pos.Sub(m.playerPosBeforeVelocityApply()).Len(){
		return
	}
	return stepPos, stepVelocity
}

func (m *Movement) pasteStateToMovements(pos, velocity mgl64.Vec3){
	m.velocity = velocity
	m.position = pos
}

func (m *Movement) doNormalCollisionThenStepAssist() (mgl64.Vec3, mgl64.Vec3){
	m.simAState(m.playerPosBeforeVelocityApply(), m.velocity)
	return m.doStepAssist()
}

func (m *Movement) simAState(pos, velocity mgl64.Vec3) (newPos, newVelocity mgl64.Vec3){
	m.stateInWorld.Position = pos
	m.stateInWorld.Velocity = velocity
	m.stateInWorld.BBoxFunc = utils.PlayerBBox
	m.stateInWorld.SimState()
	return m.stateInWorld.Position, m.stateInWorld.Velocity
}