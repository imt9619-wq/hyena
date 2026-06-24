package physics

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/utils"
)

type StateInWorld struct{
    velocity mgl64.Vec3
    position mgl64.Vec3
    aaBB     cube.BBox

    world      *blockmap.BlockMap
    scratch    *phyScratch
}

func NewStateInWorld(world *blockmap.BlockMap) *StateInWorld{
	s := &StateInWorld{}
	s.world = world
	s.scratch = newScratch()
	return s
}

type InPhyState struct{
	Velocity mgl64.Vec3
    Position mgl64.Vec3
	BBoxFunc utils.BBoxFunc
}

// we are going to round off player position to the last five digit as the player might be stuck(rare but possible) 
// if they got something like Z: 88.19999694824219 and is in front of a stair
func (s *StateInWorld) copyInPhyState(state InPhyState){
	s.position = utils.RoundVecTo5Decimal(state.Position)
	s.aaBB = state.BBoxFunc(s.position)
	s.velocity = utils.RemoveDeltaEpsilon(state.Velocity)
}

type OutPhyState struct{
	Velocity mgl64.Vec3
	Position mgl64.Vec3
	AABB     cube.BBox
}

func (s *StateInWorld) outPhyState() OutPhyState{
	return OutPhyState{
		Velocity: s.velocity,
		Position: s.position,
		AABB: s.aaBB,
	}
}

// simState will use the given aabb, velocity, position and simulate the newState which the newState will 
// replace the old one
func (s *StateInWorld) SimState(state InPhyState) OutPhyState{
	s.copyInPhyState(state)
	s.getOffset()
	s.simOffset()
	return s.outPhyState()
}

func (s *StateInWorld) simOffset(){
	if utils.DeltaIsZero(s.velocity){
		return
	}
	offsets := s.scratch.offsets.offsetArr()
	var radio float64
	deltas := s.velocity
	for axis, minRadio := range utils.MinOffset(offsets, deltas){
		radio = minRadio
		if radio != 1{
			s.velocity[axis] = 0	
		}	
	}

	s.position = s.position.Add(deltas.Mul(radio))
}

func (s *StateInWorld) getOffset(){
	if utils.DeltaIsZero(s.velocity){
		return
	}
	
	for axis := range s.velocity{
		if s.isHittingBlockOnAxis(axis){
			s.velocity[axis] = 0
		}
	}
	bm := s.world
	s.scratch.offsets.reset(s.velocity)
	for pos, model := range s.scratch.SweptBlockModels(s.aaBB, s.velocity, bm) {
		for _, blockBox := range utils.BBoxes(model, pos, bm) {
			s.scratch.offsets.considerOffsets(s.aaBB, blockBox, s.velocity)
		}
	}
}

func (s *StateInWorld) isHittingBlockOnAxis(axis int) bool{
	if s.velocity[axis] == 0{
		return false
	}
	return utils.BBoxIntersectsSolid(s.world, utils.TinyBBoxOnBBoxFace(s.aaBB, utils.FaceOnDeltaAxis(s.velocity, axis)))
}
