package physics

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/utils"
)

type StateInWorld struct{
    Velocity mgl64.Vec3
    Position mgl64.Vec3
	BBoxFunc utils.BBoxFunc
    AABB     cube.BBox

    world      *blockmap.BlockMap
    scratch    *phyScratch
}

func NewStateInWorld(world *blockmap.BlockMap) *StateInWorld{
	s := &StateInWorld{}
	s.world = world
	s.scratch = newScratch()
	return s
}

// simState will use the given aabb, velocity, position and simulate the newState which the newState will 
// replace the old one
func (s *StateInWorld) SimState(){
	s.getOffset()
	s.simOffset()
}

// we are going to round off player position to the last five digit as the player might be stuck(rare but possible) 
// if they got something like Z: 88.19999694824219 and is in front of a stair
func (s *StateInWorld) roundOffPos(){
	s.Position = utils.RoundVecTo5Decimal(s.Position)
	s.AABB = s.BBoxFunc(s.Position)
}

func (s *StateInWorld) simOffset(){
	if utils.DeltaIsZero(s.Velocity){
		return
	}
	offsets := s.scratch.offsets.offsetArr()
	var radio float64
	deltas := s.Velocity
	for axis, minRadio := range utils.MinOffset(offsets, deltas){
		radio = minRadio
		if radio != 1{
			s.Velocity[axis] = 0	
		}	
	}

	s.Position = s.Position.Add(deltas.Mul(radio))
}

func (s *StateInWorld) getOffset(){
	if utils.DeltaIsZero(s.Velocity){
		return
	}
	s.roundOffPos()
	for axis := range s.Velocity{
		if s.isHittingBlockOnAxis(axis){
			s.Velocity[axis] = 0
		}
	}
	bm := s.world
	s.scratch.offsets.reset(s.Velocity)
	for pos, model := range s.scratch.SweptBlockModels(s.AABB, s.Velocity, bm) {
		for _, blockBox := range utils.BBoxes(model, pos, bm) {
			s.scratch.offsets.considerOffsets(s.AABB, blockBox, s.Velocity)
		}
	}
}

func (s *StateInWorld) isHittingBlockOnAxis(axis int) bool{
	if s.Velocity[axis] == 0{
		return false
	}
	return utils.BBoxIntersectsSolid(s.world, utils.TinyBBoxOnBBoxFace(s.AABB, utils.FaceOnDeltaAxis(s.Velocity, axis)))
}

func (s *StateInWorld) Scratch() *phyScratch{
	return s.scratch
}