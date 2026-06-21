package physics

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/imt9619-wq/hyena/game/blockmap"
	"github.com/imt9619-wq/hyena/utils"
)

type StateInWorld struct{
    Velocity mgl64.Vec3
    Position mgl64.Vec3
    AABB     cube.BBox

    moveVector [3]int
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
	s.setMoveVector()
	s.getOffset()
	s.simOffset()
}

func (s *StateInWorld) setMoveVector(){
	for axis, plane := range s.Velocity{
		if plane == 0{
			s.moveVector[axis] = 0
		}else if plane > 0{
			s.moveVector[axis] = 1
		}else{
			s.moveVector[axis] = -1
		}
	}
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
	for axis := range s.Velocity{
		if s.isHittingWallOnAxis(axis){
			fmt.Printf("Is hitting on wall axis %v\n", axis)
			s.Velocity[axis] = 0
		}
	}
	bm := s.world
	s.scratch.offsets.reset(s.Velocity)
	for pos, model := range s.scratch.SweptBlockModels(s.AABB, s.Velocity, bm) {
		blockBoxes := utils.BBoxes(model, pos, bm)
		a := utils.DeltaAxisFace(s.Velocity)
		xSolid ,ySolid, zSolid := model.FaceSolid(pos, a[0], bm), model.FaceSolid(pos, a[1], bm), 
								  model.FaceSolid(pos, a[2], bm)
		for _, blockBox := range blockBoxes {
			s.scratch.offsets.considerOffsets(s.AABB, blockBox, [3]bool{xSolid, ySolid, zSolid}, s.Velocity)
		}
	}
}

func (s *StateInWorld) isHittingWallOnAxis(axis int) bool{
	if axis == 1 || s.Velocity[axis] == 0{
		return false
	}
	halfHori := utils.HoriProbeOffset / 2
	axisAABBpos := s.Position
	axisAABBpos[axis] += (utils.PlayerWidth/2 + halfHori) * float64(s.moveVector[axis])
	return utils.BBoxIntersectsSolid(s.world, cube.Box(
			axisAABBpos[0]-halfHori,
			axisAABBpos[1]+s.AABB.Height(),
			axisAABBpos[2]-halfHori,
			axisAABBpos[0]+halfHori,
			axisAABBpos[1],
			axisAABBpos[2]+halfHori,
			).Stretch(cube.Axis((axis/2)+1), utils.PlayerWidth/2-halfHori))
}

func (s *StateInWorld) Scratch() *phyScratch{
	return s.scratch
}