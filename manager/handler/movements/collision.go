package movements

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/imt9619-wq/hyena/utils"
)

const stepHeight = 0.6

func (m *Movement) simCollision(){
	m.doNormalCollision()
	m.doStepAssist()
	m.pasteStateToMovements()
}

func (m *Movement) doStepAssist(){
	if !m.onGround || m.velocity[1] != 0 || utils.DeltaIsZero(m.velocity){
		return
	}

	sw :=  m.stateInWorld
	var stepHeight float64 = 0
	var ceilHeight float64 = 1000
	pBBoxInStairs := utils.PlayerBBox(sw.Position)
	pBBoxInStairs = pBBoxInStairs.Extend(m.velocity.Mul(utils.HoriProbeOffset/m.velocity.Len()))
	pBBoxInStairs = pBBoxInStairs.ExtendTowards(cube.FaceUp, utils.MaxStepHeight)

	for pos, model := range utils.SweptModelsInBBox(pBBoxInStairs, m.state.BlockMap()){
		for _, blockBox := range model.BBox(pos, m.state.BlockMap()){
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
	if stepHeight > utils.MaxStepHeight || ceilHeight - stepHeight < sw.AABB.Height(){
		return
	}
}

func (m *Movement) pasteStateToMovements(){
	m.velocity = m.stateInWorld.Velocity 
	m.position = m.stateInWorld.Position
}

func (m *Movement) doNormalCollision(){
	m.stateInWorld.Position = m.playerPosBeforeVelocityApply()
	m.stateInWorld.Velocity = m.velocity
	m.stateInWorld.AABB = utils.PlayerBBox(m.playerPosBeforeVelocityApply())
	m.stateInWorld.SimState()
}
